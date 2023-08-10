package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	gflags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var opts struct {
	ServerAddr    string `short:"s" long:"server" description:"Configure the TCP Addr IP:Port to listen in Server Mode"`
	ClientAddr    string `short:"c" long:"client" description:"Configure the TCP Addr IP:Port to connect to in Client Mode"`
	UdpListenAddr string `short:"u" long:"udp" description:"Configure the UDP Addr IP:Port to listen for incoming packets"`
	UdpDestAddr   string `short:"d" long:"dest" description:"Configure the UDP Addr IP:Port to sent incoming packets to"`
	Verbose       bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func main() {
	_, err := gflags.Parse(&opts)

	if err != nil {
		// fmt.Println(err)
		return
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "02.01.2006 - 15:04:05",
	})

	if opts.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if opts.ServerAddr != "" && opts.ClientAddr != "" {
		logrus.Fatal("Error: Both server (-s) and client (-c) modes are selected.")
	}

	if opts.ServerAddr == "" && opts.ClientAddr == "" {
		logrus.Fatal("Error: One of server (-s) or client (-c) modes must be selected.")
		return
	}

	// For simplicity, just showing the selected mode.
	if opts.ServerAddr != "" {
		logrus.Info("Running in server mode on ", opts.ServerAddr)
		runServerMode()
	} else {
		logrus.Info("Running in client mode, connecting to ", opts.ClientAddr)
		runClientMode()
	}
}

func runServerMode() {
	// Open TCP listener on the given port.
	tcpAddr, err := net.ResolveTCPAddr("tcp", opts.ServerAddr)
	if err != nil {
		logrus.Fatal("Error resolving TCP address:", err)
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logrus.Fatal("Error setting up TCP listener:", err)
	}
	defer tcpListener.Close()

	udpAddr, err := net.ResolveUDPAddr("udp", opts.UdpListenAddr)
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpDestAddr, err := net.ResolveUDPAddr("udp", opts.UdpDestAddr)
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
	}

	udpConn, err := net.DialUDP("udp", udpAddr, udpDestAddr)
	if err != nil {
		logrus.Fatal("Error dialing UDP from ", udpAddr, " to ", udpDestAddr, ":", err)
	}
	defer udpConn.Close()

	for {
		logrus.Info("Waiting for incoming TCP connections on port", tcpAddr.Port)
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			logrus.Info("Error accepting TCP connection, trying again... Error:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		logrus.Info("TCP client connected:", tcpConn.RemoteAddr())
		defer tcpConn.Close()

		// Start relaying data
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			relayUDPtoTCP(udpConn, tcpConn)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			go relayTCPtoUDP(tcpConn, udpConn)
			wg.Done()
		}()

		wg.Wait()
		tcpConn.Close()
		time.Sleep(1 * time.Second)
	}

}

func runClientMode() {
	// Connect to the specified TCP server.
	tcpAddr, err := net.ResolveTCPAddr("tcp", opts.ClientAddr)
	if err != nil {
		logrus.Fatal("Error resolving TCP address:", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", opts.UdpListenAddr)
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
		return
	}

	udpDestAddr, err := net.ResolveUDPAddr("udp", opts.UdpDestAddr)
	if err != nil {
		logrus.Fatal("Error resolving UDP address:", err)
		return
	}

	udpConn, err := net.DialUDP("udp", udpAddr, udpDestAddr)
	if err != nil {
		logrus.Fatal("Error dialing UDP from ", udpAddr, " to ", udpDestAddr, ":", err)
	}
	defer udpConn.Close()

	// Retry in case connection does not work
	for {

		tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			logrus.Info("Error connecting to TCP server, trying again... Error:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		logrus.Info("Connected to TCP server:", tcpConn.RemoteAddr())

		// Start relaying data
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			relayUDPtoTCP(udpConn, tcpConn)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			go relayTCPtoUDP(tcpConn, udpConn)
			wg.Done()
		}()

		wg.Wait()
		tcpConn.Close()
		time.Sleep(1 * time.Second)
	}

}
func relayUDPtoTCP(udpConn *net.UDPConn, tcpConn *net.TCPConn) error {
	buffer := make([]byte, 4096) // Some buffer size to read the UDP packets
	lengthBytes := make([]byte, 2)

	for {

		n, _, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			logrus.Debug("Error reading from UDP:", err)
			continue
		}

		// Packet format: first 2 bytes denote the length
		lengthBytes[0] = byte(n >> 8)   // High byte
		lengthBytes[1] = byte(n & 0xFF) // Low byte

		// We only fail in case the tcp connection breaks, so we need to reconnect
		_, err = tcpConn.Write(lengthBytes) // Send the length first
		if err != nil {
			return fmt.Errorf("Error writing length to  TCP: %s", err)
		}
		_, err = tcpConn.Write(buffer[:n]) // Send the actual data
		if err != nil {
			return fmt.Errorf("Error writing data to  TCP: %s", err)
		}
	}
}

func relayTCPtoUDP(tcpConn *net.TCPConn, udpConn *net.UDPConn) error {
	buffer := make([]byte, 4096) // Some buffer size to read the TCP packets

	for {
		// First, read the length of the packet
		_, err := io.ReadFull(tcpConn, buffer[:2])
		if err != nil {
			return fmt.Errorf("Error reading length from TCP: %s", err)
		}
		packetLength := int(buffer[0])<<8 + int(buffer[1])

		// Read the actual packet data
		_, err = io.ReadFull(tcpConn, buffer[:packetLength])
		if err != nil {
			return fmt.Errorf("Error reading data from TCP: %s", err)
		}

		_, err = udpConn.Write(buffer[:packetLength]) // Send data as UDP
		if err != nil {
			logrus.Debug("Error writing to UDP:", err)
		}
	}
}
