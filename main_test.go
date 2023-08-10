package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestBridge(t *testing.T) {
	// Run server
	opts.ServerAddr = "127.0.0.1:40000"
	opts.UdpDestAddr = "127.0.0.1:40001"
	opts.UdpListenAddr = "127.0.0.1:40000"

	go func() {
		runServerMode()
	}()

	time.Sleep(1 * time.Second)

	// Run server
	opts.ClientAddr = "127.0.0.1:40000"
	opts.UdpDestAddr = "127.0.0.1:30001"
	opts.UdpListenAddr = "127.0.0.1:30000"

	go func() {
		runClientMode()
	}()

	// Wait, this is a bit odd but okay for now...
	time.Sleep(1 * time.Second)

	udpAddrApp1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40001")
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpAddrBridge1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpConn1, err := net.DialUDP("udp", udpAddrApp1, udpAddrBridge1)
	if err != nil {
		logrus.Fatal("Error dialing UDP from ", udpAddrApp1, " to ", udpAddrBridge1, ":", err)
	}

	udpAddrApp2, err := net.ResolveUDPAddr("udp", "127.0.0.1:30001")
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpAddrBridge2, err := net.ResolveUDPAddr("udp", "127.0.0.1:30000")
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpConn2, err := net.DialUDP("udp", udpAddrApp2, udpAddrBridge2)
	if err != nil {
		logrus.Fatal("Error dialing UDP from ", udpAddrApp2, " to ", udpAddrBridge2, ":", err)
	}

	// Send in first direction, and then in the backwards direction
	sendBuf := make([]byte, 1024)
	recvBuf := make([]byte, 1024)
	sendN := 0
	recvN := 0
	go func() {
		recvN, err = udpConn2.Read(recvBuf)
		if err != nil {
			t.Log("Failed to read on remote site: ", err)
			os.Exit(1)
		}

		if sendN != recvN {
			t.Log("Mismatching amount of bytes written/read")
			os.Exit(1)
		}

		sendN, err = udpConn2.Write(sendBuf)
		if err != nil {
			t.Log("Failed to write on remote site: ", err)
			os.Exit(1)
		}
	}()

	sendN, err = udpConn1.Write(sendBuf)
	if err != nil {
		t.Log("Failed to write on local site: ", err)
		t.FailNow()
	}

	recvN, err = udpConn1.Read(recvBuf)
	if err != nil {
		t.Log("Failed to read on local site: ", err)
		t.FailNow()
	}

	if sendN != recvN {
		t.Log("Mismatching amount of bytes written/read")
		t.FailNow()
	}
}
