package ai

import (
	"github.com/reusedev/draw-hub/internal/consts"
)

type Token struct {
	Token    string
	Desc     string
	Supplier consts.ModelSupplier
}

func (t Token) GetSupplier() consts.ModelSupplier {
	return t.Supplier
}
