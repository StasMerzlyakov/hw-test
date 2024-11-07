package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func worker(taskChan <-chan Task,
	stopCh <-chan struct{},
	doneChan chan<- struct{},
	errChan chan<- struct{},
) {
	defer func() {
		doneChan <- struct{}{}
	}()
	for task := range taskChan {
		// test stopCh
		select {
		case <-stopCh:
			return
		default:
		}

		if err := task(); err != nil {
			errChan <- struct{}{}
		}
	}
}

func waitWorkers(maxErrorsCount int,
	allDoneCount int,
	stopCh chan struct{},
	doneChan <-chan struct{},
	errChan <-chan struct{},
) int {
	errCount := 0
	doneCount := allDoneCount
	var once sync.Once
	for {
		select {
		case <-errChan:
			errCount++
			if errCount >= maxErrorsCount {
				once.Do(func() { // only once
					close(stopCh) // stop goroutines
				})
			}
		case <-doneChan: // wait workers
			doneCount--
			if doneCount == 0 {
				return errCount
			}
		}
	}
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n < 1 {
		return fmt.Errorf("wrong n value")
	}

	if m < 0 {
		m = 0 // always return ErrErrorsLimitExceeded !!
	}

	stopCh := make(chan struct{})
	taskChan := make(chan Task, len(tasks))
	errChan := make(chan struct{}, n)
	doneChan := make(chan struct{}, n)

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	for i := 0; i < n; i++ {
		// start workers
		go func() {
			worker(taskChan, stopCh, doneChan, errChan)
		}()
	}

	errCount := waitWorkers(m, n, stopCh, doneChan, errChan)

	if errCount >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
