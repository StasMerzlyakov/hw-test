package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func worker(taskChan <-chan Task, stopCh <-chan struct{}, doneChan chan<- struct{}, errChan chan<- struct{}) {
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
	errCount := 0

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		doneCount := n
		defer wg.Done()

		var once sync.Once
		for {
			select {
			case <-errChan:
				errCount++
				if errCount >= m {
					once.Do(func() { // only once
						close(stopCh) // stop goroutines
					})
				}
			case <-doneChan: // wait workers
				doneCount--
				if doneCount == 0 {
					return
				}
			}
		}
	}()

	for i := 0; i < n; i++ {
		// start workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(taskChan, stopCh, doneChan, errChan)
		}()
	}

	wg.Wait()
	if errCount >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
