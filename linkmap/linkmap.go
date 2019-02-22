package linkmap

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Jille/bindlink/multiplexer"
)

type Map struct {
	mp    *multiplexer.Mux
	links map[*net.UDPAddr]int
}

func New(mp *multiplexer.Mux) *Map {
	return &Map{
		mp:    mp,
		links: map[*net.UDPAddr]int{},
	}
}

func (lm *Map) StartListener(port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	sock, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	go lm.handleSocket(sock)
	return nil
}

func (lm *Map) InitiateLink(proxyAddr string) error {
	addr, err := net.ResolveUDPAddr("udp", proxyAddr)
	if err != nil {
		return err
	}
	sock, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	go lm.handleSocket(sock)
	return nil
}

func (lm *Map) Run() {
	for {
		time.Sleep(time.Second)
		cp := lm.mp.CraftControl()
		// TODO: broadcast
		_ = cp
	}
}

func (lm *Map) handleSocket(s *net.UDPConn) {
	buf := make([]byte, 8192)
	for {
		n, addr, err := s.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}
		lm.handlePacket(addr, buf[:n])
	}
}

func (lm *Map) handlePacket(addr *net.UDPAddr, buf []byte) {
}

func (lm *Map) Send(link int, packet []byte) error {
	return nil
}
