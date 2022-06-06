package gocoin

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"sync"
)

type Task struct {
	Err     error
	ExecFn  func(ctx context.Context, acc GenericAccount) (AccountActivity, error)
	Account GenericAccount
}

func NewTask(f func(ctx context.Context, acc GenericAccount) (AccountActivity, error), account GenericAccount) *Task {
	return &Task{ExecFn: f, Account: account}
}

func (t *Task) Run(ctx context.Context, acc GenericAccount) (AccountActivity, error) {
	fmt.Printf("Worker calling ExecFn for account %s \n", acc.ID )
	return t.ExecFn(ctx, acc)
}

func (wp WorkerPool) GenerateFrom(tasks []*Task) {
	for _, task := range tasks {
		wp.tasksChan <- task
	}
	close(wp.tasksChan)
}

type WorkerPool struct {
	concurrency int
	tasksChan   chan *Task
	results     chan AccountActivity
	Done         chan struct{}
}

func NewWorkerPool(concurrency int) WorkerPool {

	return WorkerPool{
		concurrency: concurrency,
		tasksChan:   make(chan *Task, 400),
		results:     make(chan AccountActivity, 400),
		Done:        make(chan struct{}),
	}
}

func (wp *WorkerPool) Run(ctx context.Context)  {
	var wg sync.WaitGroup

	for n := 0; n < wp.concurrency; n++ {
		wg.Add(1)
		go worker(ctx, wp.tasksChan, wp.results, &wg)
	}

	wg.Wait()
	close(wp.Done)
	close(wp.results)
}

func (wp WorkerPool) ResultSet() <-chan AccountActivity {
	return wp.results
}

func worker(ctx context.Context, tasksChan <-chan *Task, results chan<- AccountActivity, wg *sync.WaitGroup) {
	defer wg.Done()

	loop:
	for {
		select {
		case task, ok := <-tasksChan:
			if !ok {
				logger.Info("Channel is no longer open", zap.Bool("Taskchan status", ok))
				break loop
			}
			res , _ := task.Run(ctx, task.Account)
			results <- res
			continue loop
		case <-ctx.Done():
			logger.Info("Shutting down worker. Error detail: ", zap.Error(ctx.Err()))
			break loop
		}
	}
}
