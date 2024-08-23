package main

import (
	"net/netip"
	"testing"
)

func TestMaskCidr(t *testing.T) {
	maskCidr(netip.AddrFrom4([4]byte{1, 1, 1, 1}))
}
