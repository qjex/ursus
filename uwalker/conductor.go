package main

import (
	"context"
	"log"
	"net"
	"time"
	"uwalker/scan"
)

type connectionKey struct {
	ip   string
	port uint16
}

type ConnectionState interface {
	Init() []byte
	Read(data []byte) ([]byte, int, bool)
}

type Prober interface {
	Probe(dst net.IP, port uint16) error
	ProbeData(dst net.IP, port uint16, seq, ack uint32, data []byte) error
}

type Terminator interface {
	Terminate(dst net.IP, port uint16, seq uint32) error
}

type Sender interface {
	Prober
	Terminator
}
type Limiter interface {
	Limit(nowNanos int64) bool
}

type connection struct {
	seq          uint32
	unacked      []byte
	partyNextSeq uint32

	lstPacket time.Time
	state     ConnectionState
}

func (c *connection) ack(count int) {
	if count <= 0 {
		return // nothing new acked
	}
	rem := len(c.unacked)
	newLen := rem - count
	c.seq += uint32(count)
	if newLen <= 0 {
		c.unacked = nil
		return
	}
	nw := make([]byte, newLen)
	copy(nw, c.unacked[0:newLen])
	c.unacked = nw
}

type Conductor struct {
	ips   <-chan net.IP
	ports []uint16

	packets <-chan *scan.Packet

	s Sender
	l Limiter

	connections  map[connectionKey]*connection
	stateBuilder func() ConnectionState

	txQ chan *txReq
}

func NewConductor(
	ips <-chan net.IP,
	ports []uint16,
	packets <-chan *scan.Packet,
	s Sender,
	l Limiter,
	stateBuilder func() ConnectionState,
) *Conductor {

	return &Conductor{
		ips:          ips,
		ports:        ports,
		packets:      packets,
		s:            s,
		l:            l,
		stateBuilder: stateBuilder,

		connections: make(map[connectionKey]*connection),
		txQ:         make(chan *txReq),
	}
}

type txReq struct {
	data []byte
	seq  uint32
	ack  uint32

	addr net.IP
	port uint16
}

func (c *Conductor) Transmit(ctx context.Context) error {
	inits := make(chan struct {
		net.IP
		uint16
	})
	go func() {
		defer close(inits)
		for ip := range c.ips {
			for _, port := range c.ports {
				inits <- struct {
					net.IP
					uint16
				}{ip, port}
			}
		}
	}()
	defer log.Println("Stopping transmitter")
loop:
	for {
		select { // prioritize connections handling over connection init
		case req := <-c.txQ:
			c.send(func() error {
				return c.s.ProbeData(req.addr, req.port, req.seq, req.ack, req.data)
			})
			continue
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		select {
		case req, ok := <-inits:
			if !ok {
				break loop
			}
			c.send(func() error {
				return c.s.Probe(req.IP, req.uint16)
			})
			continue
		default:
		}
	}
	for { // waiting for remaining tcp connections
		select {
		case req := <-c.txQ:
			c.send(func() error {
				return c.s.ProbeData(req.addr, req.port, req.seq, req.ack, req.data)
			})
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Second):
			return nil
		}
	}
}

func (c *Conductor) send(sender func() error) {
	for {
		if c.l.Limit(time.Now().UnixNano()) {
			if err := sender(); err != nil {
				log.Println(err)
				time.Sleep(100 * time.Millisecond)
			}
			break
		}
	}
}

func (c *Conductor) terminate(p *scan.Packet, k connectionKey, conn *connection) {
	//var termSeq uint32 = 0
	if conn != nil {
		//termSeq = conn.seq
		delete(c.connections, k)
	}
	_ = c.s.Terminate(p.Addr, k.port, p.Ack)
}

func (c *Conductor) Collect() {
	for p := range c.packets {
		k := connectionKey{
			p.Addr.String(),
			p.Port,
		}
		conn := c.connections[k]
		res := c.handle(p, k, conn)
		if res == nil {
			c.terminate(p, k, conn)
			continue
		}
		c.txQ <- res
	}
}

func (c *Conductor) handle(p *scan.Packet, k connectionKey, conn *connection) *txReq {
	if p.Done {
		return nil
	}
	if conn == nil {
		if !p.Start {
			// Connection state lost or deleted
			return nil
		}
		conn = &connection{
			0,
			nil,
			p.Seq,
			time.Now(),
			c.stateBuilder(),
		}
		c.connections[k] = conn
	} else {
		if p.Start {
			return conn.prepRes(p)
		}
	}
	acked := p.Ack - conn.seq
	conn.ack(int(acked))
	if conn.partyNextSeq < p.Seq {
		log.Printf("Reordering detected for %s:%d", k.ip, k.port)
		return conn.prepRes(p)
	}
	if conn.partyNextSeq > p.Seq {
		log.Printf("Unexpected party seq for %s:%d. Killing the connection", k.ip, k.port)
		return nil
	}
	var res []byte
	var read int
	var finished bool
	if p.Start {
		res = conn.state.Init()
	} else if p.Data != nil {
		res, read, finished = conn.state.Read(p.Data)
		if finished {
			log.Printf("!!!Connection finished %s:%d!!!", p.Addr.String(), p.Port)
			return nil
		}
		if res == nil && read == 0 {
			return nil
		}
	}
	conn.unacked = append(conn.unacked, res...)

	if p.Start {
		read++ // syn contributes to the sequence
	}
	toAck := p.Seq + uint32(read)
	conn.partyNextSeq = toAck
	return conn.prepRes(p)
}

func (c *connection) prepRes(p *scan.Packet) *txReq {
	return &txReq{
		data: c.unacked,
		addr: p.Addr,
		port: p.Port,
		ack:  c.partyNextSeq,
		seq:  c.seq,
	}
}
