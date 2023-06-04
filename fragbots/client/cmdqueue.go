package client

import "github.com/Prince/fragbots/logging"

// CmdQueue is a custom queue that allows
// the client to process multiple invites concurrently

type CmdQueue struct {
	queueChannel chan func()
	quit         chan bool
}

func newCmdQueue() *CmdQueue {
	cq := &CmdQueue{
		queueChannel: make(chan func(), 100),
		quit:         make(chan bool),
	}
	return cq
}

func (cq *CmdQueue) GetTotalQueuedTasks() int {
	return len(cq.queueChannel)
}

// Start creates a new thread and processes
// the tasks one by one until the channel is closed
func (cq *CmdQueue) start() {
	go func() {
		for {
			select {
			case <-cq.quit:
				return
			case task := <-cq.queueChannel:
				if len(cq.quit) > 0 {
					return
				}
				task()
			}

		}
	}()
}

func (cq *CmdQueue) clear() {

}

func (cq *CmdQueue) stop() {
	cq.quit <- true
	close(cq.queueChannel)
	logging.Log("Stopped cmd queue")
}

func (cq *CmdQueue) addTask(task func()) {
	cq.queueChannel <- task
	return
}
