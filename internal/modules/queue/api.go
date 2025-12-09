package queue

import (
	"context"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"sync"
)

var ImageTaskQueue = NewTaskQueue(100)
var closeOnce sync.Once

func exeImageTask(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		select {
		case task, ok := <-ImageTaskQueue:
			if ok {
				wg.Add(1)
				go func() {
					task.Execute(ctx)
					wg.Done()
				}()
			} else {
				// channel close
				wg.Done()
				return
			}
		case <-ctx.Done():
			closeOnce.Do(func() {
				close(ImageTaskQueue)
				logs.Logger.Info().Msg("Image task queue closed")
			})
		}
	}
}

func InitImageTaskQueue(ctx context.Context, wg *sync.WaitGroup) {
	go exeImageTask(ctx, wg)
}
