package ai

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/reusedev/draw-hub/internal/consts"
	"sync"
	"time"
)

type TokenWithModel struct {
	Token
	Model string // supplier model
}

type Client struct {
	Id       string
	TryIndex [][]int
}

func (c *Client) CanTry(i, j int) bool {
	return c.TryIndex[i][j] == 0
}

func (c *Client) FirstRequest() bool {
	for _, row := range c.TryIndex {
		for _, column := range row {
			if column != 0 {
				return false
			}
		}
	}
	return true
}

type TokenManager struct {
	BanSupplier []consts.ModelSupplier
	ExpiredAt   []time.Time
	Token       [][]TokenWithModel
	Lock        *sync.Mutex

	Client []*Client
}

// GTokenManager [classification(slow|fast|gemini-2.5-flash-image-hd)]TokenManager
var GTokenManager map[string]TokenManager

func InitTokenManager(ctx context.Context, cla []string, tokens [][][]TokenWithModel) error {
	if len(cla) != len(tokens) {
		return fmt.Errorf("init token manager error")
	}
	GTokenManager = make(map[string]TokenManager)
	for i := 0; i < len(cla)-1; i++ {
		GTokenManager[cla[i]] = TokenManager{
			Token: tokens[i],
			Lock:  &sync.Mutex{},
		}
	}
	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-t.C:
				for _, manager := range GTokenManager {
					manager.tidy()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (t *TokenManager) getToken(clientId string) *TokenWithModel {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	var client *Client
	for _, v := range t.Client {
		if v.Id == clientId {
			client = v
			break
		}
	}
	if client == nil {
		client = &Client{Id: clientId, TryIndex: make([][]int, len(t.Token))}
		for i := range client.TryIndex {
			client.TryIndex[i] = make([]int, len(t.Token[i]))
		}
		t.Client = append(t.Client, client)
	}

	token := t.getValidToken(client)
	if token != nil {
		return token
	}
	if client.FirstRequest() {
		t.popBanSupplierIfAllBan()
		token := t.getValidToken(client)
		return token
	}
	return nil
}

func (t *TokenManager) popBanSupplierIfAllBan() {
	var hasValidToken bool
	for _, tokens := range t.Token {
		for _, token := range tokens {
			if t.validToken(token) {
				hasValidToken = true
				break
			}
		}
	}
	if !hasValidToken {
		t.BanSupplier = t.BanSupplier[1:]
		t.ExpiredAt = t.ExpiredAt[1:]
	}
}

func (t *TokenManager) getValidToken(client *Client) *TokenWithModel {
	for i, tokens := range t.Token {
		for j, token := range tokens {
			if client.CanTry(i, j) && t.validToken(token) {
				client.TryIndex[i][j] = 1
				return &token
			}
		}
	}
	return nil
}

func (t *TokenManager) GetToken(ctx context.Context, consumeSignal chan struct{}) chan TokenWithModel {
	tokenCh := make(chan TokenWithModel)
	clientId := uuid.NewString()
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(tokenCh)
				return
			default:
				token := t.getToken(clientId)
				if token == nil {
					close(tokenCh)
					return
				}
				tokenCh <- *token
				<-consumeSignal
			}
		}
	}()
	return tokenCh
}

func (t *TokenManager) validToken(token TokenWithModel) bool {
	for _, supplier := range t.BanSupplier {
		if token.GetSupplier() == supplier {
			return false
		}
	}
	return true
}

func (t *TokenManager) tidy() {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	for i := len(t.ExpiredAt) - 1; i >= 0; i-- {
		if t.ExpiredAt[i].Before(time.Now()) {
			t.BanSupplier = append(t.BanSupplier[:i], t.BanSupplier[i+1:]...)
			t.ExpiredAt = append(t.ExpiredAt[:i], t.ExpiredAt[i+1:]...)
		}
	}
}
