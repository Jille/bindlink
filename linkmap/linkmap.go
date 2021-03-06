package linkmap

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Jille/bindlink/multiplexer"
)

type UDPLikeConn interface {
	Write(b []byte) (int, error)
	ReadFromUDP(b []byte) (int, *net.UDPAddr, error)
}

var _ UDPLikeConn = &net.UDPConn{}

type Map struct {
	mtx        sync.Mutex
	mp         *multiplexer.Mux
	listener   *net.UDPConn
	nextLinkId int
	linkToAddr map[int]*net.UDPAddr
	linkToConn map[int]UDPLikeConn
}

func New(mp *multiplexer.Mux) *Map {
	return &Map{
		mp:         mp,
		linkToAddr: map[int]*net.UDPAddr{},
		linkToConn: map[int]UDPLikeConn{},
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
	lm.listener = sock
	go lm.handleSocket(-1, sock)
	return nil
}

func (lm *Map) InitiateLink(targetAddr string) error {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return err
	}
	sock, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	lm.newLink(sock, addr)
	return nil
}

func (lm *Map) InitiateLinkOverSOCKS(proxyAddr, target string) error {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()
	sock, err := NewUDPOverSocks(proxyAddr, target)
	if err != nil {
		return err
	}
	lm.newLink(sock, nil)
	return nil
}

func (lm *Map) newLink(sock UDPLikeConn, addr *net.UDPAddr) {
	lm.nextLinkId++
	linkId := lm.nextLinkId
	log.Printf("InitiateLink(%s): got link id %d", addr, linkId)
	if linkId > 255 {
		panic("ran out of link ids")
	}
	lm.mp.AddLink(linkId)
	lm.linkToConn[linkId] = sock
	lm.linkToAddr[linkId] = addr
	go lm.handleSocket(linkId, sock)
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
	buf[2] = 'C'
	copy(buf[4:], cp)
	for linkId := range lm.linkToAddr {
		buf[3] = byte(linkId)
		lm.send(linkId, buf)
	}
}

func (lm *Map) handleSocket(linkId int, sock UDPLikeConn) {
	buf := make([]byte, 8192)
	for {
		n, addr, err := sock.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				continue
			}
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}
		lm.handlePacket(linkId, sock, addr, buf[:n])
	}
}

func (lm *Map) handlePacket(linkId int, sock UDPLikeConn, addr *net.UDPAddr, buf []byte) {
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
		log.Printf("Got packet for new link %d from %s", remoteLinkId, addr)
		lm.mp.AddLink(remoteLinkId)
	}
	if linkId == -1 {
		lm.linkToAddr[remoteLinkId] = addr
		lm.linkToConn[remoteLinkId] = sock
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

func (lm *Map) send(linkId int, packet []byte) error {
	addr := lm.linkToAddr[linkId]
	sock := lm.linkToConn[linkId]
	if sock == nil {
		panic(fmt.Errorf("didn't find socket for link %d", linkId))
	}
	var err error
	if sock == lm.listener {
		_, err = lm.listener.WriteToUDP(packet, addr)
	} else {
		_, err = sock.Write(packet)
	}
	if err != nil {
		if strings.Contains(err.Error(), "no buffer space available") {
			return nil
		}
		return err
	}
	return nil
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
