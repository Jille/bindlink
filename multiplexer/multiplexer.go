package multiplexer

import (
	"bytes"
	"encoding/gob"
	"log"
	"math"

	"github.com/Jille/bindlink/multiplexer/sampler"
	"github.com/Jille/bindlink/multiplexer/tallier"
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
	sampler        *sampler.Sampler
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

func (m *Mux) pickLinks() []int {
	if m.sampler == nil {
		for id, _ := range m.links {
			return []int{id}
		}
		return []int{}
	}

	prob := float64(0)
	lut := map[int]bool{}

	// TODO add sampler.SampleDistinct to do this properly and efficiently
	for i := 0; i < 10; i++ {
		id := m.sampler.Sample()
		if _, ok := lut[id]; ok {
			continue
		}
		lut[id] = true
		prob += m.links[id].rate
		if prob > 0.99 {
			break
		}
	}

	ret := []int{}
	for id, _ := range lut {
		ret = append(ret, id)
	}
	return ret
}

func (m *Mux) Send(packet []byte) error {
	ids := m.pickLinks()
	ok := false
	var err error
	for _, id := range ids {
		err = m.sendToLink(id, packet)
		if err != nil {
			ok = true
			m.links[id].sent.Tally()
		}
	}
	if ok {
		return nil
	}
	return err
}

func (m *Mux) Received(linkId int, packet []byte) error {
	m.links[linkId].received.Tally()
	return m.sendToSystem(packet)
}

func (m *Mux) AddLink(linkId int) {
	m.links[linkId] = NewLinkStats()
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

	weights := map[int]float64{}
	for id, received := range packet.Received {
		link := m.links[id]
		sent := float64(link.sent.Count())
		if received == 0 {
			if sent == 0 {
				link.rate = float64(1)
			} else {
				link.rate = 0
			}
		} else {
			link.rate = sent / float64(received)
		}
		weights[id] = math.Pow(math.Min(1.0, link.rate), 10.)
		log.Printf(" %d: rate: %f", id, link.rate)
	}
	m.sampler = sampler.New(weights)
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
