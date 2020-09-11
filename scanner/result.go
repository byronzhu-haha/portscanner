package scanner

type result struct {
	ip    string
	port  int
	state PortState
}

type resultGroup struct {
	ip           string
	openedPorts  []int
	closedPorts  []int
	unknownPorts []int
}

type ResultManager map[string]*resultGroup

func newResultManager() ResultManager {
	return map[string]*resultGroup{}
}

func (m ResultManager) addResult(result result) {
	var (
		res, ok = m[result.ip]
	)
	if !ok {
		res = &resultGroup{
			ip: result.ip,
		}
	}
	switch result.state {
	case PortOpened:
		res.openedPorts = append(res.openedPorts, result.port)
	case PortClosed:
		res.closedPorts = append(res.closedPorts, result.port)
	case PortUnknown:
		res.unknownPorts = append(res.unknownPorts, result.port)
	}
}
