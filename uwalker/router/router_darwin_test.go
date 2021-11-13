package router

import (
	"net"
	"testing"
)

func TestDarwinRouter_Route(t *testing.T) {
	type args struct {
		dst net.IP
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{dst: net.IPv4(1, 1, 1, 1)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := DarwinRouter{}
			_, _, _, err := r.Route(tt.args.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Route() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
