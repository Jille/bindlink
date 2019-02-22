package linkmap

import (
	"time"

	"github.com/Jille/bindlink/multiplexer"
)

type Map struct {
}

func New() *Map {
	return &Map{}
}

func (lm *Map) StartListener(port int) error {
	go func() {}()
	return nil
}

func (lm *Map) InitiateLink(proxyAddr string) {
}

func (lm *Map) Run(mp *multiplexer.Mux) {
	for {
		time.Sleep(time.Second)
		cp := mp.CraftControl()
		// TODO: broadcast
		_ = cp
	}
}
