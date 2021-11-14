package scan

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/pkg/errors"
	"log"
	"net"
	"time"
)

type Packet struct {
	Addr net.IP
	Port uint16

	Done, Start bool
	Seq         uint32
	Ack         uint32
	Data        []byte
}

type scanner struct {
	iface        *net.Interface
	gw, src      net.IP
	routerHwaddr net.HardwareAddr

	srcPort uint16

	handle *pcap.Handle

	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer

	tcpTemplate
}

type tcpTemplate struct {
	eth layers.Ethernet
	ip4 layers.IPv4
}

const scannerSrcPort = 55324

func NewScanner(ip net.IP, router routing.Router) (*scanner, error) {
	s := &scanner{
		opts: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		srcPort: scannerSrcPort,
		buf:     gopacket.NewSerializeBuffer(),
	}
	iface, gw, src, err := router.Route(ip)
	if err != nil {
		return nil, err
	}
	s.gw, s.src, s.iface = gw, src, iface

	var handle *pcap.InactiveHandle
	if handle, err = pcap.NewInactiveHandle(iface.Name); err != nil {
		return nil, err
	}
	if err := handle.SetPromisc(true); err != nil {
		return nil, err
	}
	if err := handle.SetTimeout(5 * time.Second); err != nil {
		return nil, err
	}
	if err := handle.SetSnapLen(65536); err != nil {
		return nil, err
	}
	if s.handle, err = handle.Activate(); err != nil {
		return nil, err
	}
	if err = s.handle.SetBPFFilter(fmt.Sprintf("(tcp and dst port %d) or arp", scannerSrcPort)); err != nil {
		return nil, errors.Wrap(err, "error compiling incoming packets filter")
	}
	routerHwaddr, err := s.getHwAddr(gw)
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining the MAC of the router %s", gw.String())
	}
	s.routerHwaddr = routerHwaddr
	s.tcpTemplate = createTemplate(s.iface.HardwareAddr, s.routerHwaddr, s.src)
	return s, nil
}

func createTemplate(srcHw, dstHw net.HardwareAddr, src net.IP) tcpTemplate {
	return tcpTemplate{
		eth: layers.Ethernet{
			SrcMAC:       srcHw,
			DstMAC:       dstHw,
			EthernetType: layers.EthernetTypeIPv4,
		},
		ip4: layers.IPv4{
			SrcIP:    src,
			Version:  4,
			TTL:      64,
			Protocol: layers.IPProtocolTCP,
		},
	}
}

func (s *scanner) applyTemplate(dst net.IP) *tcpTemplate {
	s.tcpTemplate.ip4.DstIP = dst.To4()
	return &s.tcpTemplate
}

func (s *scanner) Terminate(dst net.IP, port uint16, seq uint32) error {
	t := s.applyTemplate(dst)
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(s.srcPort),
		DstPort: layers.TCPPort(port),
		Seq:     seq,
		RST:     true,
	}
	tcp.SetNetworkLayerForChecksum(&s.ip4)
	if err := s.send(&t.eth, &t.ip4, &tcp); err != nil {
		return errors.Wrap(err, "error sending RST")
	}
	return nil
}

func (s *scanner) Probe(dst net.IP, port uint16) error {
	t := s.applyTemplate(dst)
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(s.srcPort),
		DstPort: layers.TCPPort(port),
		Window:  200,
		SYN:     true,
	}
	tcp.SetNetworkLayerForChecksum(&s.ip4)
	if err := s.send(&t.eth, &t.ip4, &tcp); err != nil {
		return errors.Wrap(err, "error sending SYN")
	}
	return nil
}

func (s *scanner) ProbeData(dst net.IP, port uint16, seq, ack uint32, data []byte) error {
	t := s.applyTemplate(dst)
	tcp := layers.TCP{
		SrcPort: layers.TCPPort(s.srcPort),
		DstPort: layers.TCPPort(port),
		Window:  200,
		Seq:     seq,
		Ack:     ack,
		ACK:     true,
		PSH:     true,
	}
	tcp.SetNetworkLayerForChecksum(&s.ip4)
	if err := s.send(&t.eth, &t.ip4, &tcp, gopacket.Payload(data)); err != nil {
		return errors.Wrap(err, "error sending probe with data")
	}
	return nil
}

func (s *scanner) Packets(ctx context.Context) <-chan *Packet {
	out := make(chan *Packet)
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
			data, _, err := s.handle.ZeroCopyReadPacketData()
			if err == pcap.NextErrorTimeoutExpired {
				continue
			} else if err != nil {
				log.Printf("error reading packet: %v", err)
				continue
			}

			// Parse the packet.  We'd use DecodingLayerParser here if we
			// wanted to be really fast.
			packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)

			// Find the packets we care about, and print out logging
			// information about them.  All others are ignored.
			if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer == nil {
				// non ip packet
			} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
				// not a tcp
			} else if ip, ok := ipLayer.(*layers.IPv4); !ok {
				panic("ip layer is not ip")
			} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
				panic("tcp layer is not tcp")
			} else if tcp.DstPort != scannerSrcPort {
				panic("unexpected dst port")
			} else {
				var tcpPayload []byte
				if len(tcp.Payload) > 0 {
					tcpPayload = make([]byte, len(tcp.Payload))
					copy(tcpPayload, tcp.Payload)
				}
				out <- &Packet{
					ip.SrcIP,
					uint16(tcp.SrcPort),
					tcp.RST || tcp.FIN,
					tcp.SYN && tcp.ACK,
					tcp.Seq,
					tcp.Ack,
					tcpPayload,
				}
			}
		}
		log.Println("receiver closed")
		close(out)
	}()
	return out
}

func (s *scanner) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(s.buf, s.opts, l...); err != nil {
		return err
	}
	c := make([]byte, len(s.buf.Bytes()))
	copy(c, s.buf.Bytes())
	return s.handle.WritePacketData(c)
}

func (s *scanner) getHwAddr(arpIpDst net.IP) (net.HardwareAddr, error) {
	start := time.Now()
	if s.gw != nil {
		arpIpDst = s.gw
	}
	// Prepare the layers to send for an ARP request.
	eth := layers.Ethernet{
		SrcMAC:       s.iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.iface.HardwareAddr),
		SourceProtAddress: []byte(s.src.To4()),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(arpIpDst.To4()),
	}
	// Send a single ARP request packet
	if err := s.send(&eth, &arp); err != nil {
		return nil, err
	}
	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*5 {
			return nil, errors.New("timeout getting ARP reply")
		}
		data, _, err := s.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return nil, err
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			if net.IP(arp.SourceProtAddress).Equal(arpIpDst) {
				return arp.SourceHwAddress, nil
			}
		}
	}
}
