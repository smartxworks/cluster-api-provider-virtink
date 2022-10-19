package iprange

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strings"
)

type IPRange interface {
	List() []net.IP
}

type ipRange struct {
	start net.IP
	end   net.IP
}

var (
	reRange = regexp.MustCompile("^[0-9a-f.:-]+$")           // addr | addr-addr
	reCIDR  = regexp.MustCompile("^[0-9a-f.:]+/[0-9]{1,3}$") // addr/prefix_length
)

func Parse(s string) (IPRange, error) {
	var r IPRange
	switch {
	case reRange.MatchString(s):
		r = parseRange(s)
	case reCIDR.MatchString(s):
		r = parseCIDR(s)
	}

	if r == nil {
		return nil, fmt.Errorf("ip range '%s' invalid syntax", s)
	}
	return r, nil
}

func (i ipRange) List() []net.IP {
	ips := []net.IP{}
	ip := make(net.IP, len(i.start))
	copy(ip, i.start)
	for {
		if !i.contains(ip) {
			break
		}
		ips = append(ips, ip)
		ip = nextIP(ip)
	}
	return ips
}

func (i ipRange) contains(ip net.IP) bool {
	return bytes.Compare(ip, i.start) >= 0 && bytes.Compare(ip, i.end) <= 0
}

func parseRange(s string) IPRange {
	var start, end net.IP
	if idx := strings.IndexByte(s, '-'); idx != -1 {
		start, end = net.ParseIP(s[:idx]), net.ParseIP(s[idx+1:])
	} else {
		start, end = net.ParseIP(s), net.ParseIP(s)
	}

	return ipRange{start: start, end: end}
}

func parseCIDR(s string) IPRange {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		return nil
	}

	var lastIP net.IP
	for i := 0; i < len(network.IP); i++ {
		lastIP = append(lastIP, network.IP[i]|^network.Mask[i])
	}

	return parseRange(fmt.Sprintf("%s-%s", network.IP.String(), lastIP))
}

func nextIP(ip net.IP) net.IP {
	nextIP := make(net.IP, len(ip))
	copy(nextIP, ip)
	for j := len(nextIP) - 1; j >= 0; j-- {
		nextIP[j]++
		if nextIP[j] > 0 {
			break
		}
	}
	return nextIP
}
