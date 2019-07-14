package check

import (
	"context"
	log "github.com/sirupsen/logrus"
	"testing"
	"uwalker/gen"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{})
}

func TestPoolChecker_Start(t *testing.T) {
	ips := []string{
		"194.187.164.0/19",
	}
	cfg := &Config{
		TestHost: "google.com:80",
		Ports:    []int{8080},
		Workers:  5000,
	}
	s5 := Checker(&Socks5Checker{})
	p := NewPoolChecker(s5, gen.NewGenerator(ips), cfg)
	for range p.Start(context.Background()) {

	}

}
