// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin freebsd linux netbsd openbsd

package ipv4_test

import (
	"code.google.com/p/go.net/ipv4"
	"net"
	"os"
	"testing"
)

func TestReadWriteUnicastIPPayloadUDP(t *testing.T) {
	c, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.ListenPacket failed: %v", err)
	}
	defer c.Close()

	dst, err := net.ResolveUDPAddr("udp4", c.LocalAddr().String())
	if err != nil {
		t.Fatalf("net.ResolveUDPAddr failed: %v", err)
	}
	p := ipv4.NewPacketConn(c)
	cf := ipv4.FlagTTL | ipv4.FlagDst | ipv4.FlagInterface
	for i, toggle := range []bool{true, false, true} {
		if err := p.SetControlMessage(cf, toggle); err != nil {
			t.Fatalf("ipv4.PacketConn.SetControlMessage failed: %v", err)
		}
		writeThenReadPayload(t, i, p, []byte("HELLO-R-U-THERE"), dst)
	}
}

func TestReadWriteUnicastIPPayloadICMP(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("must be root")
	}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Fatalf("net.ListenPacket failed: %v", err)
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatalf("ResolveIPAddr failed: %v", err)
	}
	p := ipv4.NewPacketConn(c)
	cf := ipv4.FlagTTL | ipv4.FlagDst | ipv4.FlagInterface
	for i, toggle := range []bool{true, false, true} {
		wb, err := (&icmpMessage{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmpEcho{
				ID: os.Getpid() & 0xffff, Seq: i + 1,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}).Marshal()
		if err != nil {
			t.Fatalf("icmpMessage.Marshal failed: %v", err)
		}
		if err := p.SetControlMessage(cf, toggle); err != nil {
			t.Fatalf("ipv4.PacketConn.SetControlMessage failed: %v", err)
		}
		rb := writeThenReadPayload(t, i, p, wb, dst)
		m, err := parseICMPMessage(rb)
		if err != nil {
			t.Fatalf("parseICMPMessage failed: %v", err)
		}
		if m.Type != ipv4.ICMPTypeEchoReply || m.Code != 0 {
			t.Fatalf("got type=%v, code=%v; expected type=%v, code=%v", m.Type, m.Code, ipv4.ICMPTypeEchoReply, 0)
		}
	}
}

func TestReadWriteUnicastIPDatagram(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("must be root")
	}

	c, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Fatalf("net.ListenPacket failed: %v", err)
	}
	defer c.Close()

	dst, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatalf("ResolveIPAddr failed: %v", err)
	}
	r, err := ipv4.NewRawConn(c)
	if err != nil {
		t.Fatalf("ipv4.NewRawConn failed: %v", err)
	}
	cf := ipv4.FlagTTL | ipv4.FlagDst | ipv4.FlagInterface
	for i, toggle := range []bool{true, false, true} {
		wb, err := (&icmpMessage{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmpEcho{
				ID: os.Getpid() & 0xffff, Seq: i + 1,
				Data: []byte("HELLO-R-U-THERE"),
			},
		}).Marshal()
		if err != nil {
			t.Fatalf("icmpMessage.Marshal failed: %v", err)
		}
		if err := r.SetControlMessage(cf, toggle); err != nil {
			t.Fatalf("ipv4.RawConn.SetControlMessage failed: %v", err)
		}
		rb := writeThenReadDatagram(t, i, r, wb, nil, dst)
		m, err := parseICMPMessage(rb)
		if err != nil {
			t.Fatalf("parseICMPMessage failed: %v", err)
		}
		if m.Type != ipv4.ICMPTypeEchoReply || m.Code != 0 {
			t.Fatalf("got type=%v, code=%v; expected type=%v, code=%v", m.Type, m.Code, ipv4.ICMPTypeEchoReply, 0)
		}
	}
}
