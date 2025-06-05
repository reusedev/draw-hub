package queue

import (
	"context"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"sync"
)

var ImageTaskQueue = NewTaskQueue(100)

func exeImageTask(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		select {
		case task, ok := <-ImageTaskQueue:
			if ok {
				wg.Add(1)
				go func() {
					err := task.Execute(ctx, wg)
					if err != nil {
						logs.Logger.Err(err).Msg("Image task execution failed")
					}
				}()
			} else {
				wg.Done()
				return
			}
		case <-ctx.Done():
			close(ImageTaskQueue)
			logs.Logger.Info().Msg("Image task queue closed")
		}
	}
}

func InitImageTaskQueue(ctx context.Context, wg *sync.WaitGroup) {
	go exeImageTask(ctx, wg)
}
