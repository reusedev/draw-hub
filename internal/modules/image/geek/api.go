package geek

import (
	"github.com/reusedev/draw-hub/internal/modules/image"
)

type Geek struct {
	client *image.Client
}

func NewGeek(token string) *Geek {
	return &Geek{
		client: image.NewClient(),
	}
}

func (g *Geek) Image1() {

}
