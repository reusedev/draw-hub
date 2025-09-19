package image

type AsyncProvider interface {
	Create() AsyncAckResponse
	Query() PollResponse
}
