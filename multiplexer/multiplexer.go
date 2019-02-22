package multiplexer

type Mux struct {
	links      map[int]*Stats
	nextLinkId int
	// callback gives packets received from the tunnel to the local system
	callback func([]byte) error
}

type Stats struct {
}

func New() *Mux {
	return &Mux{
		links: map[int]*Stats{},
	}
}

func (m *Mux) Start(callback func([]byte) error) {
	m.callback = callback
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
