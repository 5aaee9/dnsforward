package main

import "net/netip"

func maskCidr(addr netip.Addr, block int) netip.Addr {
	p, err := addr.Prefix(block)
	if err != nil {
		panic(err)
	}

	return p.Masked().Addr()
}

const maskIPv4Length = 24
const maskIPv6Length = 56

func clientSubNetMask(addr netip.Addr) netip.Addr {
	if addr.Is4() {
		return maskCidr(addr, maskIPv4Length)
	}

	return maskCidr(addr, maskIPv6Length)
}
