package linkmap

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Jille/bindlink/multiplexer"
)

type Map struct {
	mtx        sync.Mutex
	mp         *multiplexer.Mux
	nextLinkId int
	addrToLink map[string]int
	linkToAddr map[int]*net.UDPAddr
	addrToSock map[string]*net.UDPConn
}

func New(mp *multiplexer.Mux) *Map {
	return &Map{
		mp:         mp,
		addrToLink: map[string]int{},
		linkToAddr: map[int]*net.UDPAddr{},
		addrToSock: map[string]*net.UDPConn{},
	}
}

func (lm *Map) StartListener(port int) error {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
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
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
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
	lm.addrToLink[addr.String()] = linkId
	lm.linkToAddr[linkId] = addr
	lm.addrToSock[addr.String()] = sock
	go lm.handleSocket(linkId, sock)
	return nil
}

func (lm *Map) Run() {
	for {
		time.Sleep(time.Second)
		lm.broadcastControl()
	}
}

func (lm *Map) broadcastControl() {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
	cp := lm.mp.CraftControl()
	buf := make([]byte, len(cp)+4)
	buf[0] = 'B'
	buf[1] = 'L'
	buf[2] = 'D'
	copy(buf[4:], cp)
	for linkId := range lm.linkToAddr {
		buf[3] = byte(linkId)
		lm.send(linkId, buf)
	}
}

func (lm *Map) handleSocket(linkId int, sock *net.UDPConn) {
	buf := make([]byte, 8192)
	for {
		n, addr, err := sock.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}
		lm.handlePacket(linkId, sock, addr, buf[:n])
	}
}

func (lm *Map) handlePacket(linkId int, sock *net.UDPConn, addr *net.UDPAddr, buf []byte) {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
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
	if _, known := lm.linkToAddr[remoteLinkId]; !known {
		lm.mp.AddLink(remoteLinkId)
		lm.addrToLink[addr.String()] = remoteLinkId
		lm.addrToSock[addr.String()] = sock
	}
	if linkId == -1 {
		lm.linkToAddr[remoteLinkId] = addr
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

func (lm *Map) send(link int, packet []byte) error {
	addr := lm.linkToAddr[link]
	_, err := lm.addrToSock[addr.String()].WriteToUDP(packet, addr)
	return err
}

func (lm *Map) Send(link int, packet []byte) error {
	buf := make([]byte, len(packet)+4)
	buf[0] = 'B'
	buf[1] = 'L'
	buf[2] = 'D'
	buf[3] = byte(link)
	copy(buf[4:], packet)
	return lm.send(link, buf)
}
