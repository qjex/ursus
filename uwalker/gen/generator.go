package gen

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"net"
)

type Generator struct {
	cidrs   []string
	blacked *sSet
}

func NewGenerator(cidrs []string, blacked []string) (*Generator, error) {
	blackedNets := make([]*net.IPNet, len(blacked))
	for i, b := range blacked {
		_, n, err := net.ParseCIDR(b)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse CIDR from the blacklist %s", b)
		}
		blackedNets[i] = n
	}
	return &Generator{
		cidrs:   cidrs,
		blacked: newSSet(blackedNets),
	}, nil
}

func (g *Generator) Ips(ctx context.Context) chan net.IP {
	out := make(chan net.IP)
	go func() {
		defer close(out)
		for _, cidr := range g.cidrs {
			ip, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Println("Error generating ip: ", err)
			}
			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
				if g.blacked.contains(ip) {
					continue
				}
				dup := make(net.IP, len(ip))
				copy(dup, ip)
				select {
				case <-ctx.Done():
					return
				case out <- dup:
				}
			}
		}
	}()
	return out
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
