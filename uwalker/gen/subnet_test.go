package gen

import (
	"net"
	"testing"
)

func Test_newSSet(t *testing.T) {
	type args struct {
		subnets []*net.IPNet
		ip      net.IP
	}
	exSnets := []*net.IPNet{
		snet("240.0.0.0/4"),
		snet("10.0.0.0/8"),
		snet("255.255.255.255/32"),
		snet("192.175.48.0/24"),
		snet("100.64.0.0/10"),
		snet("192.0.0.0/8"),
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"same prefix",
			args{
				subnets: []*net.IPNet{
					snet("192.168.0.234/32"),
					snet("192.168.0.234/24"),
					snet("192.168.0.234/16"),
				},
				ip: net.ParseIP("192.168.22.44"),
			},
			true,
		},
		{"broadcast",
			args{exSnets, net.ParseIP("255.255.255.255")},
			true,
		},
		{"private",
			args{exSnets, net.ParseIP("10.129.255.23")},
			true,
		},
		{"public",
			args{exSnets, net.ParseIP("11.129.255.23")},
			false,
		},
		{"public2",
			args{exSnets, net.ParseIP("190.0.0.0")},
			false,
		},
		{"private multiple subnets",
			args{exSnets, net.ParseIP("192.175.48.1")},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := newSSet(tt.args.subnets)
			if got := ss.contains(tt.args.ip); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func snet(cidr string) *net.IPNet {
	_, res, _ := net.ParseCIDR(cidr)
	return res
}
