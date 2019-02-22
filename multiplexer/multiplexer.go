package multiplexer

type Mux struct {
	links        map[int]*Stats
	nextLinkId   int
	sendToSystem func([]byte) error
	sendToLink   func(int, []byte) error
}

type Stats struct {
}

func New() *Mux {
	return &Mux{
		links: map[int]*Stats{},
	}
}

func (m *Mux) Start(toSystem func([]byte) error, toLink func(int, []byte) error) {
	m.sendToSystem = toSystem
	m.sendToLink = toLink
}

func (m *Mux) Send(packet []byte) error {
	return nil
}

func (m *Mux) AddLink() int {
	m.nextLinkId++
	id := m.nextLinkId
	m.links[id] = &Stats{}
	return id
}

func (m *Mux) HandleControl(linkId int, packet []byte) {
}

func (m *Mux) CraftControl() []byte {
	return nil
}
