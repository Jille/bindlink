package linkmap

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Jille/bindlink/multiplexer"
)

type Map struct {
	mp         *multiplexer.Mux
	nextLinkId int
	ipToLink   map[*net.UDPAddr]int
	linkToIP   map[int]*net.UDPAddr
}

func New(mp *multiplexer.Mux) *Map {
	return &Map{
		mp:       mp,
		ipToLink: map[*net.UDPAddr]int{},
		linkToIP: map[int]*net.UDPAddr{},
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
	go lm.handleSocket(-1, sock)
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
	lm.nextLinkId++
	linkId := lm.nextLinkId
	if linkId > 255 {
		panic("ran out of link ids")
	}
	lm.mp.AddLink(linkId)
	lm.ipToLink[addr] = linkId
	lm.linkToIP[linkId] = addr
	go lm.handleSocket(linkId, sock)
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

func (lm *Map) handleSocket(linkId int, s *net.UDPConn) {
	buf := make([]byte, 8192)
	for {
		n, addr, err := s.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}
		lm.handlePacket(linkId, addr, buf[:n])
	}
}

func (lm *Map) handlePacket(linkId int, addr *net.UDPAddr, buf []byte) {
	if len(buf) < 4 {
		log.Printf("Received short packet from %s", addr)
		return
	}
	if buf[0] != 'B' || buf[1] != 'L' {
		log.Printf("Received packet with wrong header from %s", addr)
		return
	}
	remoteLinkId := int(buf[3])
	if linkId != -1 && remoteLinkId != linkId {
		panic(fmt.Errorf("got packet for link %d over link %d", remoteLinkId, linkId))
	}
	if _, known := lm.linkToIP[remoteLinkId]; !known {
		lm.mp.AddLink(remoteLinkId)
		lm.ipToLink[addr] = remoteLinkId
	}
	if linkId == -1 {
		lm.linkToIP[remoteLinkId] = addr
	}
	switch buf[2] {
	case 'C':
		lm.mp.HandleControl(remoteLinkId, buf[4:])
	case 'D':
		lm.mp.Received(remoteLinkId, buf[4:])
	default:
		log.Printf("Packet of unknown type %q/%d from %s", buf[2], buf[2], addr)
		return
	}
}

func (lm *Map) Send(link int, packet []byte) error {
	return nil
}
