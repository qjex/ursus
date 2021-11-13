package gen

import (
	"net"
	"sort"
)

func toInt(n net.IP) uint32 {
	var res uint32 = 0
	for i := 0; i < 4; i++ {
		res <<= 8
		res |= uint32(n[i])
	}
	return res
}

// sSet stores subnets and allows checking an IP for intersection with any subnet
type sSet struct {
	subnets []*net.IPNet
}

func newSSet(subnets []*net.IPNet) *sSet {
	sort.Slice(subnets, func(i, j int) bool {
		return toInt(subnets[i].IP) < toInt(subnets[j].IP)
	})
	filtered := make([]*net.IPNet, 0)
	for _, s := range subnets {
		if len(filtered) == 0 {
			filtered = append(filtered, s)
			continue
		}
		lastI := len(filtered) - 1
		last := filtered[lastI]
		lastSz, _ := last.Mask.Size()
		curSz, _ := s.Mask.Size()
		if last.IP.Equal(s.IP) && curSz < lastSz {
			filtered[lastI] = s
			continue
		}
		if last.Contains(s.IP) {
			continue
		}
		filtered = append(filtered, s)
	}

	return &sSet{
		subnets: filtered,
	}
}

func (s *sSet) contains(ip net.IP) bool {
	ipInt := toInt(ip.To4())
	cur := sort.Search(len(s.subnets), func(i int) bool {
		return toInt(s.subnets[i].IP) > ipInt
	}) - 1
	return cur >= 0 && s.subnets[cur].Contains(ip)
}
