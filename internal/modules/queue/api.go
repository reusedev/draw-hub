package queue

import (
	"context"
	"sync"
)

var ImageTaskQueue = NewTaskQueue(100)

type Scheduler struct {
	ctx    context.Context
	queue  TaskQueue
	worker chan struct{}
	wg     *sync.WaitGroup
}

func NewQueueScheduler(ctx context.Context, queue TaskQueue, workerNum int, wg *sync.WaitGroup) *Scheduler {
	return &Scheduler{
		ctx:    ctx,
		queue:  queue,
		worker: make(chan struct{}, workerNum),
		wg:     wg,
	}
}

func (s *Scheduler) schedule() {
	for {
		select {
		case s.worker <- struct{}{}:
			select {
			case task := <-s.queue:
				go func() {
					task.Execute(s.ctx)
					<-s.worker
				}()
			default:
				<-s.worker
			}
		case <-s.ctx.Done():
			for task := range s.queue {
				task.Abort(s.ctx)
			}
			s.wg.Done()
		}
	}
}

func InitTaskScheduler(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	scheduler := NewQueueScheduler(ctx, ImageTaskQueue, 50, wg)
	go scheduler.schedule()
}
