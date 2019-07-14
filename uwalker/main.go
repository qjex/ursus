package main

import (
	"bufio"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"syscall"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{})
}

func readCIDR(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	rLimit.Cur = rLimit.Max
	if runtime.GOOS == "darwin" && rLimit.Cur > 10240 {
		// The max file limit is 10240, even though
		// the max returned by Getrlimit is 1<<63-1.
		// This is OPEN_MAX in sys/syslimits.h.
		rLimit.Cur = 10240
	}

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	log.Infof("Current fd limit: %d", rLimit.Cur)

	var c check.Config
	conf := config.NewConfig()
	conf.Load(file.NewSource(file.WithPath("config.json")))
	conf.Scan(&c)
	ips, err := readCIDR(c.CDIRfile)
	if err != nil {
		log.Fatal(err)
	}
	s5 := check.Checker(&check.Socks5Checker{})
	pool := check.NewPoolChecker(s5, gen.NewGenerator(ips), &c)

	for proxy := range pool.Start(context.Background()) {
		log.Info(proxy)
	}
}
