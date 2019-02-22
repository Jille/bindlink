// +build notun

// This is a fake implementation that just tunnels TCP rather than a full interface.
package tundev

import (
	"flag"
	"log"
	"net"
)

var (
	connectTcp = flag.String("connect_tcp", "", "Connect to this host:port and multiplex all received traffic over the tunnel")
)

type Device struct {
	conn net.Conn
}

func New() (*Device, error) {
	conn, err := net.Dial("tcp", *connectTcp)
	if err != nil {
		return nil, err
	}
	return &Device{conn}, nil
}

func (d *Device) Run(sendToMultiplexer func([]byte) error) {
	buf := make([]byte, 2000)
	for {
		n, err := d.conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to read from TCP %q: %v", *connectTcp, err)
		}
		if err := sendToMultiplexer(buf[:n]); err != nil {
			log.Fatalf("Failed to send message through multiplexer: %v", err)
		}
	}
}

func (d *Device) Send(packet []byte) error {
	_, err := d.conn.Write(packet)
	return err
}
