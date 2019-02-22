package multiplexer

type Mux struct {
	links        map[int]*Stats
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

func (m *Mux) Received(linkId int, packet []byte) error {
	return m.sendToSystem(packet)
}

func (m *Mux) AddLink(linkId int) {
	m.links[linkId] = &Stats{}
}

func (m *Mux) HandleControl(linkId int, packet []byte) {
}

func (m *Mux) CraftControl() []byte {
	return nil
}
