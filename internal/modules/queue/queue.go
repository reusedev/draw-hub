package queue

import (
	"context"
	"sync"
)

type Task interface {
	Execute(ctx context.Context, wg *sync.WaitGroup) error
	Fail(err error)
}

type TaskQueue chan Task

func NewTaskQueue(size int) TaskQueue {
	return make(TaskQueue, size)
}
