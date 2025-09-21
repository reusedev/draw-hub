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

func NewPoller(ctx context.Context, provider AsyncProvider) *Poller {
	return &Poller{
		Ctx:             ctx,
		Provider:        provider,
		RequestInterval: time.Second * 5,
	}
}

func (p *Poller) Poll(taskId int) (chan AsyncQueryResponse, chan error) {
	pollChan := make(chan AsyncQueryResponse)
	errorChan := make(chan error)
	go func() {
		t := time.NewTicker(p.RequestInterval)
		for {
			select {
			case <-t.C:
				queryResp := p.Provider.Query(taskId)
				if queryResp.GetError() != nil {
					errorChan <- queryResp.GetError()
					return
				}
				urls := queryResp.GetURLs()
				if len(urls) != 0 {
					pollChan <- queryResp
				}
			}
		}
	}()
	return pollChan, errorChan
}
