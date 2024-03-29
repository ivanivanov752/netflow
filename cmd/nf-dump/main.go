/*
Command nf-dump decodes NetFlow packets from UDP datagrams.

Usage:

	nf-dump [flags]

Flags:

	-addr string 	Listen address (default ":2055")
*/
package main

import (
	"bytes"
	"flag"
	"log"
	"net"

	"github.com/ivanivanov752/netflow"
	"github.com/ivanivanov752/netflow/ipfix"
	"github.com/ivanivanov752/netflow/netflow1"
	"github.com/ivanivanov752/netflow/netflow5"
	"github.com/ivanivanov752/netflow/netflow6"
	"github.com/ivanivanov752/netflow/netflow7"
	"github.com/ivanivanov752/netflow/netflow9"
	"github.com/ivanivanov752/netflow/session"
)

// Safe default
var readSize = 2 << 16

func main() {
	listen := flag.String("addr", ":2055", "Listen address")
	flag.Parse()

	var addr *net.UDPAddr
	var err error
	if addr, err = net.ResolveUDPAddr("udp", *listen); err != nil {
		log.Fatal(err)
	}

	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		log.Fatal(err)
	}

	if err = server.SetReadBuffer(readSize); err != nil {
		log.Fatal(err)
	}

	decoders := make(map[string]*netflow.Decoder)
	for {
		buf := make([]byte, 8192)
		var remote *net.UDPAddr
		var octets int
		if octets, remote, err = server.ReadFromUDP(buf); err != nil {
			log.Printf("error reading from %s: %v\n", remote, err)
			continue
		}

		log.Printf("received %d bytes from %s\n", octets, remote)

		d, found := decoders[remote.String()]
		if !found {
			s := session.New()
			d = netflow.NewDecoder(s)
			decoders[remote.String()] = d
		}

		m, err := d.Read(bytes.NewBuffer(buf[:octets]))
		if err != nil {
			log.Println("decoder error:", err)
			continue
		}

		switch p := m.(type) {
		case *netflow1.Packet:
			netflow1.Dump(p)

		case *netflow5.Packet:
			netflow5.Dump(p)

		case *netflow6.Packet:
			netflow6.Dump(p)

		case *netflow7.Packet:
			netflow7.Dump(p)

		case *netflow9.Packet:
			netflow9.Dump(p)

		case *ipfix.Message:
			ipfix.Dump(p)
		}
	}
}
