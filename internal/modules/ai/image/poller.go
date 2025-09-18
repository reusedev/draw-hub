package image

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/observer"
)

type Poller struct {
	Observers []observer.Observer
}

func NewPoller() {

}

func (p *Poller) Notify() {
	for _, o := range p.Observers {
		o.Update(consts.EventPoll, nil)
	}
}

type PollResponse struct {
	Status PollTaskStatus
	Urls   []string
}

type PollTaskStatus int

const (
	PollStatusSuccess PollTaskStatus = iota
	PollStatusFailed
	PollStatusRunning
)
