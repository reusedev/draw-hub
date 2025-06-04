package queue

import (
	"context"
	"sync"
)

type Task interface {
	Execute(ctx context.Context, wg *sync.WaitGroup) error
}

type TaskQueue chan Task

func NewTaskQueue(size int) TaskQueue {
	return make(TaskQueue, size)
}
