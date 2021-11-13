package gen

import (
	"context"
	"net"
	"testing"
)

func TestGenerator_Ips(t *testing.T) {
	type fields struct {
		cidrs []string
	}
	type args struct {
		ctx context.Context
	}
	blacklist := []string{"192.168.0.1/16", "10.20.30.1/24"}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int32
	}{
		{
			name: "gen ips",
			fields: fields{
				cidrs: []string{"3.83.0.0/16"},
			},
			args: struct{ ctx context.Context }{ctx: context.Background()},
			want: 65536,
		},
		{
			name: "single",
			fields: fields{
				cidrs: []string{"1.1.1.1/32"},
			},
			args: struct{ ctx context.Context }{ctx: context.Background()},
			want: 1,
		},
		{
			name: "blacklist intersection",
			fields: fields{
				cidrs: []string{"10.20.0.0/16", "192.168.10.1/24"},
			},
			args: struct{ ctx context.Context }{ctx: context.Background()},
			want: 65280,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGenerator(tt.fields.cidrs, blacklist)
			if err != nil {
				t.Fatal(err)
			}
			got := chanSz(g.Ips(tt.args.ctx))
			if int32(got) != tt.want {
				t.Errorf("Generator.Ips() = %v, want %v", got, tt.want)
			}
		})
	}
}

func chanSz(c chan net.IP) int {
	m := make(map[string]interface{})
	for i := range c {
		m[i.String()] = true
	}
	return len(m)
}
