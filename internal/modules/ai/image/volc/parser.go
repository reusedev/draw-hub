package volc

import "time"

type CreateResponse struct {
	Supplier       string        `json:"supplier"`
	TokenDesc      string        `json:"token_desc"`
	Model          string        `json:"model"`
	StatusCode     int           `json:"status_code"`
	RespBody       string        `json:"resp_body"`
	RespAt         time.Time     `json:"resp_at"`
	Duration       time.Duration `json:"duration"`
	Error          error         `json:"error,omitempty"`
	ProviderTaskID int           `json:"provider_task_id"`
	TaskID         int           `json:"task_id"`
}

func (c *CreateResponse) GetTaskID() int {
	return c.TaskID
}

func (c *CreateResponse) GetProviderTaskID() int {
	return c.ProviderTaskID
}

func (c *CreateResponse) GetError() error {
	return c.Error
}

func (c *CreateResponse) SetTaskID(taskID int) {
	c.TaskID = taskID
}

func (c *CreateResponse) SetError(err error) {
	c.Error = err
}

func (c *CreateResponse) SetProviderTaskID(ptID int) {
	c.ProviderTaskID = ptID
}

func (c *CreateResponse) SetBasicResponse(statusCode int, respBody string, respAt time.Time) {
	c.StatusCode = statusCode
	c.RespBody = respBody
	c.RespAt = respAt
}

type QueryResponse struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	URLs       []string      `json:"URLs"`
	Error      error         `json:"error,omitempty"`
	TaskID     int           `json:"task_id"`
}
