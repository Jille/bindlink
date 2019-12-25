// +build !notun

package tundev

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	"github.com/Jille/bindlink/prefbuf"
	"github.com/songgao/water"
)

var (
	mtu = flag.Int("mtu", 1460, "MTU to use for tundev")
)

type Device struct {
	ifce *water.Interface
}

func New(isMaster bool) (*Device, error) {
	ips := map[bool]string{
		true:  "10.10.10.1",
		false: "10.10.10.2",
	}
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Interface name: %s", ifce.Name())
	var c *exec.Cmd
	if runtime.GOOS == "darwin" {
		c = exec.Command(
			"ifconfig",
			ifce.Name(),
			ips[isMaster],
			ips[!isMaster],
			"mtu",
			strconv.Itoa(*mtu),
		)
	} else {
		c = exec.Command(
			"ifconfig",
			ifce.Name(),
			ips[isMaster],
			"netmask",
			"255.255.255.252",
			"mtu",
			strconv.Itoa(*mtu),
		)
	}
	if out, err := c.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s: %s", err, out)
	}
	if f, ok := ifce.ReadWriteCloser.(*os.File); ok {
		if err := syscall.SetNonblock(int(f.Fd()), false); err != nil {
			return nil, fmt.Errorf("Failed to set blocking mode: %v", err)
		}
	} else {
		log.Printf("Couldn't cast to os.File. Might crash with EAGAIN.")
	}
	return &Device{ifce}, nil
}

func (d *Device) Run(sendToMultiplexer func([]byte) error) {
	buf := prefbuf.Alloc(2000)
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
