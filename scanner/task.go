package scanner

type task struct {
	ip   string
	port int
}

type TaskManager struct {
	taskNum      int32
	ips          []string
	ports        []int
	consumerDone chan struct{}
	producerDone chan int32
	tasks        chan task
}

func newTaskManager(ips []string, ports []int, goroutines int32) *TaskManager {
	taskNum := int32(len(ips) * len(ports))
	if taskNum < goroutines {
		goroutines = taskNum
	}
	return &TaskManager{
		taskNum:      taskNum,
		ips:          ips,
		ports:        ports,
		consumerDone: make(chan struct{}),
		producerDone: make(chan int32),
		tasks:        make(chan task, goroutines),
	}
}

func (m *TaskManager) productTask() {
	for _, ip := range m.ips {
		for _, port := range m.ports {
			m.tasks <- task{
				ip:   ip,
				port: port,
			}
		}
	}
	m.producerDone <- m.taskNum
	close(m.tasks)
}

func (m *TaskManager) consumeTask(scanFunc func(ip string, port int) PortState, resultChan chan result) {
	for task := range m.tasks {
		go func(ip string, port int) {
			result := result{
				ip:   ip,
				port: port,
			}
			switch scanFunc(ip, port) {
			case PortOpened:
				result.state = PortOpened
			case PortClosed:
				result.state = PortClosed
			default:
				result.state = PortUnknown
			}
			resultChan <- result
		}(task.ip, task.port)
	}
	<-m.consumerDone
}
