package iprange_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/smartxworks/cluster-api-provider-virtink/controllers/iprange"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		addr        string
		expectCount int
		expectStart string
		expectEnd   string
	}{
		{"192.168.240.241", 1, "192.168.240.241", "192.168.240.241"},
		{"192.168.240.242-192.168.240.247", 6, "192.168.240.242", "192.168.240.247"},
		{"192.168.240.248/29", 8, "192.168.240.248", "192.168.240.255"},
		{"192.168.240.249/29", 8, "192.168.240.248", "192.168.240.255"},
	}
	for _, testCase := range testCases {
		ipRange, err := iprange.Parse(testCase.addr)
		assert.Nil(t, err)

		ips := ipRange.List()
		assert.Equal(t, testCase.expectCount, len(ips), testCase.addr)
		assert.Equal(t, testCase.expectStart, ips[0].String(), testCase.addr)
		assert.Equal(t, testCase.expectEnd, ips[len(ips)-1].String(), testCase.addr)
	}
}
