package multiplexer

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/Jille/bindlink/tallier"
)

type ControlPacket struct {
	SeqNo    int
	Received map[int]int64
}

type Mux struct {
	links          map[int]*LinkStats
	sendToSystem   func([]byte) error
	sendToLink     func(int, []byte) error
	ourCtrlSeqNo   int
	theirCtrlSeqNo int
}

type LinkStats struct {
	sent     *tallier.Tallier
	received *tallier.Tallier
	rate     float64
}

func New() *Mux {
	return &Mux{
		links: map[int]*LinkStats{},
	}
}

func (m *Mux) Start(toSystem func([]byte) error, toLink func(int, []byte) error) {
	m.sendToSystem = toSystem
	m.sendToLink = toLink
}

func (m *Mux) Send(packet []byte) error {
	return nil
}

func (m *Mux) AddLink(id int) {
	m.links[id] = NewLinkStats()
}

func (m *Mux) HandleControl(linkId int, buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	var packet ControlPacket
	if err := dec.Decode(&packet); err != nil {
		log.Printf("CraftControl: gob.Decode(): %v", err)
		return
	}

	if packet.SeqNo == m.theirCtrlSeqNo {
		return // Already seen this control packet
	}

	m.theirCtrlSeqNo = packet.SeqNo

	for id, received := range packet.Received {
		link := m.links[id]
		link.rate = float64(link.sent.Count()) / float64(received)
		log.Printf(" %d: rate: %f", id, link.rate)
	}
}

func (m *Mux) CraftControl() []byte {
	m.ourCtrlSeqNo++
	packet := ControlPacket{
		SeqNo:    m.ourCtrlSeqNo,
		Received: map[int]int64{},
	}

	for id, link := range m.links {
		packet.Received[id] = link.received.Count()
	}

	// encode
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(packet); err != nil {
		log.Fatalf("CraftControl: gob.Encode(): %v", err)
	}
	return buf.Bytes()
}

func NewLinkStats() *LinkStats {
	return &LinkStats{
		sent:     tallier.New(500, 30000), // 30s window with 500ms bucket size
		received: tallier.New(500, 30000),
	}
}
