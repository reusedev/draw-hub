package queue

import (
	"context"
)

type Task interface {
	Execute(ctx context.Context)
	Abort(ctx context.Context)
}

type TaskQueue chan Task

func NewTaskQueue(size int) TaskQueue {
	return make(TaskQueue, size)
}
