package check

import (
	"context"
	"github.com/qjex/ursus/gen"
	"golang.org/x/sync/semaphore"
	"net"
	"strconv"
	"sync"
)

type Proxy struct {
	Addr net.IP
	Port int
	Type string
}

type CheckerPool struct {
	checker   Checker
	generator *gen.Generator
	ports     []int
	testHost  string
	sem       *semaphore.Weighted
}

type Checker interface {
	Check(checkHost, testHost string) bool
	Type() string
}

func NewCheckerPool(checker Checker, generator *gen.Generator, config *Config) *CheckerPool {
	return &CheckerPool{
		checker:   checker,
		generator: generator,
		ports:     config.Ports,
		testHost:  config.TestHost,
		sem:       semaphore.NewWeighted(int64(config.Workers)),
	}
}

func (p *CheckerPool) Start(ctx context.Context) <-chan Proxy {
	out := make(chan Proxy)
	go func() {
		var wg sync.WaitGroup
		for ip := range p.generator.Ips(ctx) {
			wg.Add(1)
			go func(ip net.IP) {
				err := p.sem.Acquire(ctx, 1)
				if err == nil {
					p.worker(ip, out)
					p.sem.Release(1)
				}
				wg.Done()
			}(ip)
		}
		go func() {
			wg.Wait()
			close(out)
		}()
	}()
	return out
}

func (p *CheckerPool) worker(ip net.IP, out chan<- Proxy) {
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
