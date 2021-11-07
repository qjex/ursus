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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				cidrs: tt.fields.cidrs,
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
