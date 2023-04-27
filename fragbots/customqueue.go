package main

type FragQueue interface {
	Start()
	Stop()
	AddTask(task func())
	GetTotalQueuedTasks() int
}

type fragQueue struct {
	queueChannel chan func()
}

func NewFragQueue() FragQueue {
	fq := &fragQueue{
		queueChannel: make(chan func(), 100),
	}
	return fq
}

func (fq *fragQueue) GetTotalQueuedTasks() int {
	return len(fq.queueChannel)
}
func (fq *fragQueue) Start() {
	go func() {
		for task := range fq.queueChannel {
			task()
		}
	}()
}

func (fq *fragQueue) Stop() {
	close(fq.queueChannel)
}

func (fq *fragQueue) AddTask(task func()) {
	fq.queueChannel <- task
	return
}
