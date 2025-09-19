package image

import (
	"context"
	"time"
)

type Poller struct {
	Ctx             context.Context
	Provider        AsyncProvider
	RequestInterval time.Duration
}

func NewPoller() {

}
