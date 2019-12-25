package linkmap

import (
	"fmt"
	"log"
	"net"
	"time"
)

func setupSOCKS(proxy string, target, localAddr *net.UDPAddr) (*net.TCPConn, *net.UDPAddr, error) {
	proxyAddr, err := net.ResolveTCPAddr("tcp", proxy)
	if err != nil {
		return nil, nil, err
	}

	sock, err := net.DialTCP("tcp", nil, proxyAddr)
	if err != nil {
		return nil, nil, err
	}

	if localAddr.IP.IsUnspecified() {
		la := *localAddr
		la.IP = sock.LocalAddr().(*net.TCPAddr).IP
		localAddr = &la
	}

	var greetingReq [3]byte
	greetingReq[0] = 5 // SOCKS version
	greetingReq[1] = 1 // number of authentication methods supported
	greetingReq[2] = 0 // no authentication

	if _, err := sock.Write(greetingReq[:]); err != nil {
		return nil, nil, err
	}

	var greetingResp [2]byte
	if _, err := sock.Read(greetingResp[:]); err != nil {
		return nil, nil, err
	}

	if greetingResp[0] != 5 { // SOCKS version
		return nil, nil, fmt.Errorf("unexpected version in greeting: %d, wanted 5", greetingResp[0])
	}
	if greetingResp[1] != 0 { // chosen authentication method
		return nil, nil, fmt.Errorf("unexpected authentication method in greeting: %d, wanted 0", greetingResp[1])
	}

	var connectReq [10]byte
	connectReq[0] = 5
	connectReq[1] = 3 // UDP
	connectReq[2] = 0 // reserved
	writeHostPort(connectReq[3:], localAddr)
	fmt.Printf("connectReq: %v\n", connectReq)

	if _, err := sock.Write(connectReq[:]); err != nil {
		return nil, nil, err
	}

	var connectResp [100]byte
	n, err := sock.Read(connectResp[:])
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("Received connect response: %v\n", connectResp[:n])

	// expect
	if connectResp[0] != 5 { // SOCKS version
		return nil, nil, fmt.Errorf("unexpected version in connect: %d, wanted 5", connectResp[0])
	}
	if connectResp[1] != 0 { // "request granted"
		return nil, nil, fmt.Errorf("unexpected status in connect: %d, wanted 0", connectResp[1])
	}
	addr, err := decodeHostPort(connectResp[3:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode host in connect response: %v", err)
	}

	if addr.IP == nil || addr.IP.IsUnspecified() {
		addr.IP = proxyAddr.IP
	}

	return sock, addr, nil
}

func sizeOfHostPort(addr *net.UDPAddr) int {
	return 7
}

func writeHostPort(buf []byte, addr *net.UDPAddr) {
	if ip := addr.IP.To4(); ip != nil {
		buf[0] = 1 // ipv4 (1=IPv4 address, 3=domain name, 4=IPv6 address)
		buf[1] = ip[0]
		buf[2] = ip[1]
		buf[3] = ip[2]
		buf[4] = ip[3]
	} else if ip := addr.IP.To16(); ip != nil {
		buf[0] = 4 // ipv6 (1=IPv4 address, 3=domain name, 4=IPv6 address)
		copy(buf[1:], ip[:16])
	}
	buf[5] = byte((addr.Port >> 8) & 255)
	buf[6] = byte(addr.Port & 255)
}

func decodeHostPort(b []byte) (*net.UDPAddr, error) {
	var ip net.IP
	switch b[0] {
	case 1: // IPv4
		ip = net.IPv4(b[1], b[2], b[3], b[4])
		b = b[5:]
	case 4: // IPv6
		ip := make(net.IP, 16)
		copy(ip, b[1:17])
		b = b[17:]
	default:
		return nil, fmt.Errorf("address type should be 1 or 4, was %d", b[0])
	}
	port := (int(b[0]) << 8) + int(b[1])
	return &net.UDPAddr{IP: ip, Port: port}, nil
}

func NewUDPOverSocks(proxyAddr, targetAddr string) (*UDPOverSocks, error) {
	sock, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp4", targetAddr)
	if err != nil {
		return nil, err
	}
	u := &UDPOverSocks{
		udpConn:    sock,
		proxyAddr:  proxyAddr,
		targetAddr: addr,
	}
	go func() {
		var b [100]byte
		for {
			u.lastTCPError = u.connect()
			if u.lastTCPError != nil {
				log.Printf("Failed to connect to proxy: %v", u.lastTCPError)
				time.Sleep(time.Second)
				continue
			}
			// This should block until our connection dies.
			var n int
			n, u.lastTCPError = u.tcpConn.Read(b[:])
			log.Printf("Unexpectedly received %v (%q)", b[:n], b[:n])
			u.tcpConn.Close()
			time.Sleep(time.Second)
		}
	}()
	return u, nil
}

type UDPOverSocks struct {
	tcpConn      net.Conn
	lastTCPError error
	udpConn      *net.UDPConn
	proxyAddr    string
	targetAddr   *net.UDPAddr
	udpProxyAddr *net.UDPAddr
}

func (u *UDPOverSocks) connect() error {
	if u.tcpConn != nil {
		u.tcpConn.Close()
		u.tcpConn = nil
	}
	conn, addr, err := setupSOCKS(u.proxyAddr, u.targetAddr, u.udpConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		return err
	}
	u.tcpConn = conn
	if err = conn.SetKeepAlive(true); err != nil {
		return fmt.Errorf("SetKeepAlive: %v", err)
	}
	if err = conn.SetKeepAlivePeriod(4 * time.Second); err != nil {
		return fmt.Errorf("SetKeepAlivePeriod: %v", err)
	}
	u.udpProxyAddr = addr
	return nil
}

func (u *UDPOverSocks) Write(b []byte) (int, error) {
	buf := make([]byte, len(b)+sizeOfHostPort(u.targetAddr)+3)
	buf[0] = 0
	buf[1] = 0
	buf[2] = 0
	writeHostPort(buf[3:], u.targetAddr)
	copy(buf[3+sizeOfHostPort(u.targetAddr):], b)
	n, err := u.udpConn.WriteToUDP(buf, u.udpProxyAddr)
	n -= 10
	if n < 0 {
		n = 0
	}
	return n, err
}

func (u *UDPOverSocks) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	n, _, err := u.udpConn.ReadFromUDP(b)
	if err != nil {
		return 0, nil, err
	}
	if len(b) < 10 {
		log.Printf("Unexpected socks encapsulated UDP packet: too short")
	}
	if b[0] != 0 {
		log.Printf("Unexpected socks encapsulated UDP packet: reserved0 should be 0, was %d", b[0])
	}
	if b[1] != 0 {
		log.Printf("Unexpected socks encapsulated UDP packet: reserved1 should be 0, was %d", b[1])
	}
	if b[2] != 0 { // Fragment number
		log.Printf("Unexpected socks encapsulated UDP packet: fragment number should be 0, was %d", b[2])
	}
	addr, err := decodeHostPort(b[3:])
	if err != nil {
		log.Printf("Unexpected socks encapsulated UDP packet: %v", err)
	}
	n -= 10
	copy(b, b[10:])
	return n, addr, nil
}
