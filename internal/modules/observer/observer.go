package observer

type Observer interface {
	Update(event string, data interface{})
}

type Subject interface {
	Notify()
}
