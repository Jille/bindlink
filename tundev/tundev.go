// +build !notun

package tundev

import (
	"log"

	"github.com/songgao/water"
)

type Device struct {
	ifce *water.Interface
}

func New() (*Device, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Interface name: %s", ifce.Name())
	return &Device{ifce}, nil
}

func (d *Device) Run(sendToMultiplexer func([]byte) error) {
	buf := make([]byte, 2000)
	for {
		n, err := d.ifce.Read(buf)
		if err != nil {
			log.Fatalf("Failed to read from interface %s: %v", d.ifce.Name(), err)
		}
		if err := sendToMultiplexer(buf[:n]); err != nil {
			log.Fatalf("Failed to send message through multiplexer: %v", err)
		}
	}
}

func (d *Device) Send(packet []byte) error {
	_, err := d.ifce.Write(packet)
	return err
}
