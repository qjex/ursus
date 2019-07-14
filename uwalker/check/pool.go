package check

import (
	"context"
	"golang.org/x/sync/semaphore"
	"net"
	"strconv"
	"sync"
	"uwalker/gen"
)

type PoolChecker struct {
	checker   Checker
	generator *gen.Generator
	ports     []int
	testHost  string
	sem       *semaphore.Weighted
}

func NewPoolChecker(checker Checker, generator *gen.Generator, config *Config) *PoolChecker {
	return &PoolChecker{
		checker:   checker,
		generator: generator,
		ports:     config.Ports,
		testHost:  config.TestHost,
		sem:       semaphore.NewWeighted(int64(config.Workers)),
	}
}

func (p *PoolChecker) Start(ctx context.Context) <-chan Proxy {
	out := make(chan Proxy)
	go func() {
		var wg sync.WaitGroup
		for ip := range p.generator.Ips(ctx) {
			err := p.sem.Acquire(ctx, 1)
			if err == nil {
				wg.Add(1)
				go func(ip net.IP) {
					if err == nil {
						p.worker(ip, out)
						p.sem.Release(1)
					}
					wg.Done()
				}(ip)
			}
		}
		go func() {
			wg.Wait()
			close(out)
		}()
	}()
	return out
}

func (p *PoolChecker) worker(ip net.IP, out chan<- Proxy) {
	for _, port := range p.ports {
		proxy := ip.String() + ":" + strconv.Itoa(port)
		if p.checker.Check(proxy, p.testHost) {
			out <- Proxy{
				Addr: ip,
				Port: port,
				Type: p.checker.Type(),
			}
		}
	}
}
