package tundev

type Device struct {
}

func New() (*Device, error) {
	return &Device{}, nil
}

func (d *Device) Start(callback func([]byte) error) error {
	return nil
}

func (d *Device) Send(packet []byte) error {
	return nil
}
