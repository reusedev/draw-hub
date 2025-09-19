package observer

type Observer interface {
	Update(event int, data interface{})
}
