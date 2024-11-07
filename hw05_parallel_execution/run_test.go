package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun1(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})
}

func TestRun2(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestAdditional1(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("test_max_erros_0", func(t *testing.T) {
		tasksCount := 50
		workersCount := 5
		maxErrorsCount := 0

		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var concurrentTaskCount int32

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))

			tasks = append(tasks, func() error {
				defer atomic.AddInt32(&concurrentTaskCount, -1)
				atomic.AddInt32(&concurrentTaskCount, 1)
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		err := Run(tasks, workersCount, maxErrorsCount)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
	})
}

func TestAdditional2(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("test_eventually", func(t *testing.T) {
		tasksCount := 50
		workersCount := 5
		maxErrorsCount := 1

		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var concurrentTaskCount int32

		require.Eventually(t, func() bool {
			return concurrentTaskCount <= int32(workersCount)
		}, time.Second, 10*time.Millisecond)

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))

			tasks = append(tasks, func() error {
				defer atomic.AddInt32(&concurrentTaskCount, -1)
				atomic.AddInt32(&concurrentTaskCount, 1)
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
	})
}
