package ai

import (
	"context"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestGetToken(t *testing.T) {
	m := TokenManager{
		Token: [][]TokenWithModel{
			{
				{
					Token{Token: "sk-1"},
					"gpt-4o-image",
				},
				{
					Token{Token: "sk-2"},
					"gpt-4o-image",
				},
			},
			{
				{
					Token{Token: "sk-3"},
					"gpt-4o-image-vip",
				},
			},
		},
		Lock:   &sync.Mutex{},
		Client: make([]*Client, 0),
	}
	tokens := make([]TokenWithModel, 0)
	signal := make(chan struct{})
	for token := range m.GetToken(context.Background(), signal) {
		tokens = append(tokens, token)
		go func() {
			signal <- struct{}{}
		}()
	}
	require.Equal(t, []TokenWithModel{
		{
			Token{Token: "sk-1"},
			"gpt-4o-image",
		},
		{
			Token{Token: "sk-2"},
			"gpt-4o-image",
		},
		{
			Token{Token: "sk-3"},
			"gpt-4o-image-vip",
		}}, tokens)
}

func TestTidy(t *testing.T) {
	fiveMinLater := time.Now().Add(5 * time.Minute)
	m := TokenManager{
		BanSupplier: []consts.ModelSupplier{
			consts.Tuzi,
			consts.Geek,
			consts.V3,
		},
		ExpiredAt: []time.Time{time.Now().Add(-5 * time.Minute), fiveMinLater, time.Now().Add(-5 * time.Minute)},
		Lock:      &sync.Mutex{},
	}
	m.tidy()
	require.Equal(t, []consts.ModelSupplier{consts.Geek}, m.BanSupplier)
	require.Equal(t, []time.Time{fiveMinLater}, m.ExpiredAt)
}
