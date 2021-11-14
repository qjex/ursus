package scan

import (
	"context"
	"net"
	"testing"
	"time"
	"uwalker/limiter"
	"uwalker/router"
)

func Test_scanner_Probe(t *testing.T) {
	s := create(t)

	type args struct {
		dst  net.IP
		port uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"184.181.217.210:4145", args{net.ParseIP("184.181.217.210"), 4145}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Probe(tt.args.dst, tt.args.port); (err != nil) != tt.wantErr {
				t.Errorf("Probe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Benchmark_scanner_Probe(b *testing.B) {
	s := create(b)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		for range s.Packets(ctx) {

		}
		close(done)
	}()
	l := limiter.NewLogLimiter(1000)
	dst := net.ParseIP("110.101.153.121")
	b.Run("bench", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if l.Limit(time.Now().Unix()) {
				err := s.Probe(dst, 4145)
				if err != nil {
					b.Fatal(err, n)
				}
			}
		}
	})
	cancel()
	<-done
}

func create(t testing.TB) *scanner {
	r, err := router.New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewScanner(net.IPv4(8, 8, 8, 8), r)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
