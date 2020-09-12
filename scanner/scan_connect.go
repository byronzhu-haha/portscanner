package scanner

import (
	"github.com/byronzhu-haha/portscanner/logger"
	"net"
	"sort"
	"strings"
)

type connectScanner struct {
	hasStarted bool
	result     chan result
	resultMgr  ResultManager
	conf       Config
	taskMgr    *TaskManager
	logger     logger.Logger
}

func newConnectScanner(ips []string, ports []int, conf Config, logger logger.Logger) PortScanner {
	return &connectScanner{
		result:    make(chan result, conf.goroutines),
		resultMgr: newResultManager(),
		conf:      conf,
		taskMgr:   newTaskManager(ips, ports, conf.goroutines),
		logger:    logger,
	}
}

func (s *connectScanner) Start() {
	s.logger.Infof("scan start...")
	go s.taskMgr.productTask()
	go func() {
		var (
			count  int32
			total  int32
			result result
		)
		for {
			select {
			case result = <-s.result:
				count++
				s.resultMgr.addResult(result)
			case total = <-s.taskMgr.producerDone:
				s.logger.Infof("product tasks finished!")
				if total == count {
					s.taskMgr.consumerDone <- struct{}{}
					s.logger.Infof("scan finished!")
					return
				}
			}
			if total != 0 && count >= total {
				s.logger.Infof("consume tasks finished!")
				s.taskMgr.consumerDone <- struct{}{}
				break
			}
		}
	}()
	s.hasStarted = true
}

func (s *connectScanner) Scan() {
	if !s.hasStarted {
		panic(ErrNotStart)
	}
	s.taskMgr.consumeTask(s.connect, s.result)
}

func (s *connectScanner) connect(ip string, port int) PortState {
	conn, err := net.DialTimeout("tcp", pack(ip, port), s.conf.timeout)
	if err != nil {
		if strings.Contains(err.Error(), "refused") {
			return PortClosed
		}
		return PortUnknown
	}
	_ = conn.Close()
	return PortOpened
}

func (s *connectScanner) Stop() {
	s.logger.Infof("scanner stop...")
	s.logger.Close()
	close(s.taskMgr.consumerDone)
	close(s.taskMgr.producerDone)
	close(s.result)
	s.hasStarted = false
	s.resultMgr = nil
	s.taskMgr = nil
	s.logger = nil
}

func (s *connectScanner) Output() {
	if !s.hasStarted {
		panic(ErrNotStart)
	}
	s.logger.Infof("start print result:")
	for ip, group := range s.resultMgr {
		sort.Ints(group.openedPorts)
		sort.Ints(group.closedPorts)
		sort.Ints(group.unknownPorts)
		s.logger.Debugf(
			"ip(%s):\n\t\t\t\topened port: %+v,\n\t\t\t\tclosed port: %+v,\n\t\t\t\tunknown port: %+v.",
			ip,
			group.openedPorts,
			group.closedPorts,
			group.unknownPorts,
		)
	}
	s.logger.Infof("end print result.")
}
