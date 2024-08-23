package main

import (
	"context"
	"flag"
	"github.com/miekg/dns"
	"log"
	"net"
	"net/netip"
)

func appendEdns0Subnet(msg *dns.Msg, addr net.IP) {
	newOpt := true
	var o *dns.OPT
	for _, v := range msg.Extra {
		if v.Header().Rrtype == dns.TypeOPT {
			o = v.(*dns.OPT)
			newOpt = false
			break
		}
	}
	if o == nil {
		o = new(dns.OPT)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
	}
	e := &dns.EDNS0_SUBNET{
		Code:        dns.EDNS0SUBNET,
		SourceScope: 0,
		Address:     addr,
	}
	if addr.To4() == nil {
		e.Family = 2 // IP6
		e.SourceNetmask = maskIPv6Length
	} else {
		e.Family = 1 // IP4
		e.SourceNetmask = maskIPv4Length
	}
	o.Option = append(o.Option, e)
	if newOpt {
		msg.Extra = append(msg.Extra, o)
	}
}

func main() {
	listen := flag.String("listen", "0.0.0.0:53", "DNS server listen address")
	upstream := flag.String("upstream", "1.1.1.1:53", "Upstream dns server")
	flag.Parse()

	client := &dns.Client{}

	handler := dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		remoteAddrPort := clientSubNetMask(netip.MustParseAddrPort(w.RemoteAddr().String()).Addr())
		forwardMessage := r.Copy()

		appendEdns0Subnet(forwardMessage, net.ParseIP(remoteAddrPort.String()))

		res, _, err := client.Exchange(forwardMessage, *upstream)
		if err != nil {
			return
		}

		_ = w.WriteMsg(res)
	})

	udpServer := &dns.Server{
		Addr:    *listen,
		Net:     "udp",
		Handler: handler,
	}

	tcpServer := &dns.Server{
		Addr:    *listen,
		Net:     "tcp",
		Handler: handler,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := udpServer.ListenAndServe()
		if err != nil {
			cancel()
			log.Fatal(err)
		}
	}()

	go func() {
		err := tcpServer.ListenAndServe()
		if err != nil {
			cancel()
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
}
