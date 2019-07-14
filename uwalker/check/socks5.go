package check

import (
	"github.com/fangdingjun/socks-go"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

type Socks5Checker struct {
}

func (sc *Socks5Checker) Check(checkHost, testHost string) bool {
	log.Tracef("socks5: Checking %s", checkHost)
	conn, err := net.DialTimeout("tcp", checkHost, 5*time.Second)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			log.Warn(err)
		}
		return false
	}

	s := socks.Client{Conn: conn}
	err = s.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Trace("Error setting deadline: ", err)
		return false
	}
	hostName := strings.TrimRight(testHost, ":")
	port, _ := strconv.Atoi(strings.TrimLeft(testHost, ":"))
	err = s.Connect(hostName, uint16(port))
	if err != nil {
		log.Trace(err)
		return false
	}
	err = conn.Close()
	if err != nil {
		log.Warn(err)
	}
	return true
}

func (sc *Socks5Checker) Type() string {
	return "socks5"
}
