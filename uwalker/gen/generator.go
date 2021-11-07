package gen

import (
	"context"
	"log"
	"net"
)

type Generator struct {
	cidrs []string
}

func NewGenerator(cidrs []string) *Generator {
	return &Generator{
		cidrs: cidrs,
	}
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
