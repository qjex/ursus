//+build darwin

package router

import (
	"fmt"
	"github.com/google/gopacket/routing"
	"github.com/pkg/errors"
	"net"
	"os/exec"
	"strings"
)

func New() (routing.Router, error) {
	return DarwinRouter{}, nil
}

type DarwinRouter struct {
	ifName string
}

func (r DarwinRouter) Route(dst net.IP) (iface *net.Interface, gateway, preferredSrc net.IP, err error) {
	gw, ifName, err := discoverGateway(dst)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "error while routing the %s", dst)
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "error listing all interfaces")
	}
	for _, i := range ifaces {
		if i.Name != ifName {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "error listing all addresses")
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
			case *net.IPNet:
				ones, _ := v.Mask.Size()
				if ones > 32 {
					continue // skipping subnet
				}
				return &i, gw, v.IP, nil
			}
		}
		break
	}
	return nil, nil, nil, errors.New(fmt.Sprintf("can't determine route for %s", dst))

}

func (r DarwinRouter) RouteWithSrc(input net.HardwareAddr, src, dst net.IP) (iface *net.Interface, gateway, preferredSrc net.IP, err error) {
	panic("implement me")
}

func discoverGateway(dst net.IP) (net.IP, string, error) {
	routeCmd := exec.Command("/sbin/route", "-n", "get", dst.String())
	output, err := routeCmd.CombinedOutput()
	if err != nil {
		return nil, "", errors.Wrap(err, "error executing /sbin/route")
	}
	gw, ifName := parse(output)
	return gw, ifName, nil
}

func parse(output []byte) (gw net.IP, ifName string) {
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		header := strings.TrimLeft(fields[0], " ")
		value := fields[1]
		switch header {
		case "gateway:":
			gw = net.ParseIP(value)
		case "interface:":
			ifName = value
		}
	}

	return
}
