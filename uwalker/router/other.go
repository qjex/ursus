//+build !darwin

package router

import "github.com/google/gopacket/routing"

func New() (routing.Router, error) {
	return routing.New()
}
