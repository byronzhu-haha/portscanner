package scanner

import "strconv"

type PortScanner interface {
	Start()
	Scan()
	Stop()
	Output()
}

type TypeScanner string
type PortState byte

const (
	Connect = "connect"
	SYN     = "syn"
	FIN     = "fin"
	Device  = "device"
)

const (
	PortUnknown = iota
	PortOpened
	PortClosed
)

func NewScanner(ips []string, ports []int, opts ...Option) PortScanner {
	conf := Config{}
	for _, opt := range opts {
		conf = opt(conf)
	}
	switch conf.typ {
	case Connect:
		return newConnectScanner(ips, ports, conf)
	case SYN:
		return newSynScanner()
	case FIN:
		return newFinScanner()
	case Device:
		return newDeviceScanner()
	default:
		return newConnectScanner(ips, ports, conf)
	}
}

func pack(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}
