package control

import (
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"time"
	"ursus/store"
)

const (
	hbRcv    = 0x0
	hbAck    = 0x1
	proxyRcv = 0x2
)

const (
	socks5 = 0x1
)

var ackBuf = []byte{hbAck}

type Server interface {
	Start()
	Stop()
}

type srv struct {
	l       net.Listener
	done    chan struct{}
	tasks   chan net.Conn
	workers int
	wg      sync.WaitGroup
	store   store.Store
}

func NewServer(bind string, workers int, store store.Store) (Server, error) {
	l, err := net.Listen("tcp4", bind)
	if err != nil {
		return nil, errors.Wrap(err, "error starting listening socket")
	}
	done := make(chan struct{})
	tasks := make(chan net.Conn)
	return &srv{
		l:       l,
		done:    done,
		tasks:   tasks,
		workers: workers,
		wg:      sync.WaitGroup{},
		store:   store,
	}, nil
}

func (s *srv) Start() {
	for i := 0; i < s.workers; i++ {
		go s.worker()
	}
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				log.Printf("Error accepting: %s\n", err)
				continue
			}
			log.Printf("Accepting: %s\n", conn.RemoteAddr())
			select {
			case s.tasks <- conn:
			case <-s.done:
				s.l.Close()
				log.Println("Waiting for workers to shutdown")
				s.wg.Wait()
				close(s.tasks)
				return
			}
		}
	}()
}

func (s *srv) Stop() {
	close(s.done)
}

func (s *srv) worker() {
	for {
		select {
		case <-s.done:
			return
		case c := <-s.tasks:
			s.wg.Add(1)
			s.handle(c)
			s.wg.Done()
		}
	}
}

func (s *srv) handle(conn net.Conn) {
	defer conn.Close()
	for {
		setDeadline(conn)
		buf := make([]byte, 8)
		cnt, err := conn.Read(buf)
		if err != nil || cnt == 0 {
			log.Printf("Error reading from the socket: %s", err)
			return
		}
		switch buf[0] {
		case hbRcv:
			if cnt != 1 {
				return
			}
		case proxyRcv:
			if cnt != 8 {
				return
			}
			var proto string
			switch buf[1] {
			case socks5:
				proto = "socks5"
			default:
				log.Printf("Unknown proxy protocol %s", buf[1])
				return
			}
			ip := net.IP(buf[2:6])
			port := uint16(16*buf[6] + buf[6])
			proxy := store.Proxy{
				Addr:    ip,
				Port:    port,
				Proto:   proto,
				Updated: time.Now(),
			}
			err := s.store.Save(proxy)
			if err != nil {
				log.Printf("Error saving proxy %v, %s", proxy, err)
			}
		default:
			return
		}
		err = sendAck(conn)
		if err != nil {
			log.Printf("Error writing heartbeat %s", err)
		}

	}
}

func sendAck(conn net.Conn) error {
	setDeadline(conn)
	wcnt, err := conn.Write(ackBuf)

	if wcnt != 1 {
		return errors.New("Invalid written bytes")
	}
	return err
}

func setDeadline(conn net.Conn) {
	err := conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Printf("Couldn't set deadline: %s\n", err)
	}
}
