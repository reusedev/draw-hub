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
		case task := <-ImageTaskQueue:
			go func() {
				wg.Add(1)
				err := task.Execute(ctx, wg)
				if err != nil {
					logs.Logger.Err(err).Msg("Image task execution failed")
				}
			}()
		case <-ctx.Done():
			close(ImageTaskQueue)
			logs.Logger.Info().Msg("Image task queue stopped")
			for task := range ImageTaskQueue {
				wg.Add(1)
				if err := task.Execute(ctx, wg); err != nil {
					logs.Logger.Err(err).Msg("Shutdown task failed")
				}
			}
			wg.Done()
			return
		}
	}
}

func InitImageTaskQueue(ctx context.Context, wg *sync.WaitGroup) {
	go exeImageTask(ctx, wg)
}
