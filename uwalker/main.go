package main

import (
	"bufio"
	"context"
	_ "embed"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"uwalker/banner"
	"uwalker/gen"
	"uwalker/limiter"
	"uwalker/router"
	"uwalker/scan"
	"uwalker/storage"
)

//go:embed static/excludes
var defaultBlacklist string

func readCIDR(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

var opts struct {
	TestHost  string `long:"test-host" default:"google.com:80"`
	Subnet    string `short:"s" description:"Subnet to scan, e.g 192.168.0.1/24"`
	Cidrs     string `short:"f" description:"File with subnets to scan"`
	Ports     string `short:"p" description:"Ports to scan, e.g comma separated \"2055,2056,1999\" or ranges \"2055-2059,1999\"" required:"true"`
	BlackList string `short:"b" description:"Specifies file with excluded subnets from scanning in the same format as the subnets for scanning. If it is not specified, the default one would be used"`
	Rate      uint32 `short:"r" description:"Max probing rate in packet/s" default:"100"`
	Sqlite    string `long:"sqlite" description:"Path to the SQLite database" default:"db.sqlite"`
}

func parsePorts(p string) ([]uint16, error) {
	conv := func(r string) (uint16, error) {
		port, err := strconv.Atoi(r)
		if err != nil {
			return 0, errors.Errorf("invalid port %s format", r)
		}
		return uint16(port), nil
	}
	ports := make([]uint16, 0)
	ranges := strings.Split(p, ",")
	for _, r := range ranges {
		if !strings.Contains(r, "-") {
			// single
			port, err := conv(r)
			if err != nil {
				return nil, err
			}
			ports = append(ports, port)
			continue
		}
		// range
		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, errors.Errorf("invalid port range %s format", r)
		}
		from, err := conv(parts[0])
		if err != nil {
			return nil, err
		}
		to, err := conv(parts[1])
		if err != nil {
			return nil, err
		}
		for i := from; i <= to; i++ {
			ports = append(ports, i)
		}
	}
	return ports, nil
}

func getExcludes(blacklist string) ([]string, error) {
	if blacklist == "" {
		return readCIDR(strings.NewReader(defaultBlacklist))
	}
	return readCIDRFile(blacklist)
}

func getCIDRs(path string, subnet string) ([]string, error) {
	if subnet != "" {
		return []string{subnet}, nil
	}
	return readCIDRFile(path)
}

func readCIDRFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readCIDR(f)
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	if opts.Cidrs == "" && opts.Subnet == "" {
		println("Either subnet or file with subnets to scan must be defined. See the -h")
		os.Exit(1)
	}
	ports, err := parsePorts(opts.Ports)
	if err != nil {
		log.Fatal("failed to parse ports for scanning: ", err)
	}
	excludes, err := getExcludes(opts.BlackList)
	if err != nil {
		log.Fatal("failed to read the file with excludes: ", err)
	}
	ips, err := getCIDRs(opts.Cidrs, opts.Subnet)
	if err != nil {
		log.Fatal("failed to read the file with subnets for scanning: ", err)
	}

	gen, err := gen.NewGenerator(ips, excludes)
	if err != nil {
		log.Fatal("failed to init the tool with provided subnets: ", err)
	}
	l := limiter.NewLogLimiter(opts.Rate)
	r, err := router.New()
	if err != nil {
		log.Fatal("failed to init routing subsystem: ", err)
	}
	s, err := scan.NewScanner(net.IPv4(1, 1, 1, 1), r)
	if err != nil {
		log.Fatal(err)
	}
	engine, err := storage.NewSqlite(opts.Sqlite)
	if err != nil {
		log.Fatal(err)
	}
	store, err := storage.NewStore(engine)
	if err != nil {
		log.Fatal("failed to init the store: ", err)
	}
	c := NewConductor(ports, s, l, func() ConnectionState {
		return &banner.Socks5{}
	})

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
	established := c.Collect(s.Packets(ctx))
	go func() {
		_ = c.Transmit(gen.Ips(ctx))
		cancel()
	}()
	persist(store, established)
}

func persist(store *storage.Store, established <-chan Protocol) {
	for e := range established {
		log.Printf("protocol %s detected at the %s:%d", e.Proto, e.Ip.String(), e.Port)
		if err := store.PersistBanner(e.Ip, e.Port, e.Proto); err != nil {
			log.Println("failed to persist the banner")
		}
	}
}
