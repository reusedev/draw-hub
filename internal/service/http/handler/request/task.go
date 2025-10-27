package request

import (
	"github.com/reusedev/draw-hub/internal/consts"
)

type TaskForm interface {
	GetImageOrigin() string
	GetGroupId() string
	GetImageIds() []int
	GetModel() string
	GetPrompt() string
	GetSpeed() consts.TaskSpeed
	GetQuality() string
	GetSize() string
	GetTaskType() string
}

type SlowTask struct {
	ImageType string `form:"image_type"`
	GroupId   string `form:"group_id"`
	ImageId   int    `form:"image_id"`
	ImageIds  []int  `form:"image_ids"`
	Prompt    string `form:"prompt"`
}

func (s *SlowTask) GetImageOrigin() string {
	return s.ImageType
}
func (s *SlowTask) GetGroupId() string {
	return s.GroupId
}
func (s *SlowTask) GetImageIds() []int {
	if len(s.ImageIds) != 0 {
		return s.ImageIds
	}
	return []int{s.ImageId}
}
func (s *SlowTask) GetModel() string {
	return ""
}
func (s *SlowTask) GetPrompt() string {
	return s.Prompt
}
func (s *SlowTask) GetQuality() string {
	return ""
}
func (s *SlowTask) GetSize() string {
	return ""
}
func (s *SlowTask) GetSpeed() consts.TaskSpeed {
	return consts.SlowSpeed
}
func (s *SlowTask) GetTaskType() string {
	if len(s.ImageIds) != 0 {
		return consts.TaskTypeEdit.String()
	}
	return consts.TaskTypeGenerate.String()
}

type FastSpeed struct {
	ImageType string `form:"image_type"`
	GroupId   string `form:"group_id"`
	ImageId   int    `form:"image_id"`
	ImageIds  []int  `form:"image_ids"`
	Prompt    string `form:"prompt"`
	Quality   string `form:"quality"`
	Size      string `form:"size"`
}

func (s *FastSpeed) GetImageOrigin() string {
	return s.ImageType
}
func (s *FastSpeed) GetGroupId() string {
	return s.GroupId
}
func (s *FastSpeed) GetImageIds() []int {
	if len(s.ImageIds) != 0 {
		return s.ImageIds
	}
	return []int{s.ImageId}
}
func (s *FastSpeed) GetModel() string {
	return ""
}
func (s *FastSpeed) GetPrompt() string {
	return s.Prompt
}
func (s *FastSpeed) GetQuality() string {
	return s.Quality
}
func (s *FastSpeed) GetSize() string {
	return s.Size
}
func (s *FastSpeed) GetSpeed() consts.TaskSpeed {
	return consts.FastSpeed
}
func (s *FastSpeed) GetTaskType() string {
	if len(s.ImageIds) != 0 {
		return consts.TaskTypeEdit.String()
	}
	return consts.TaskTypeGenerate.String()
}

type Generate struct {
	GroupId string `form:"group_id"`
	Prompt  string `form:"prompt"`
}

func (g *Generate) GetImageOrigin() string {
	return ""
}
func (g *Generate) GetGroupId() string {
	return g.GroupId
}
func (g *Generate) GetImageIds() []int {
	return []int{}
}
func (g *Generate) GetModel() string {
	return ""
}
func (g *Generate) GetPrompt() string {
	return g.Prompt
}
func (g *Generate) GetQuality() string {
	return ""
}
func (g *Generate) GetSize() string {
	return ""
}
func (g *Generate) GetSpeed() consts.TaskSpeed {
	return ""
}
func (g *Generate) GetTaskType() string {
	return consts.TaskTypeGenerate.String()
}

type Create struct {
	Model     string `form:"model"`
	ImageType string `form:"image_type"`
	GroupId   string `form:"group_id"`
	ImageIds  []int  `form:"image_ids"`
	Prompt    string `form:"prompt"`
	Size      string `form:"size"`
}

func (c *Create) GetImageOrigin() string {
	return c.ImageType
}
func (c *Create) GetGroupId() string {
	return c.GroupId
}
func (c *Create) GetImageIds() []int {
	if len(c.ImageIds) != 0 {
		return c.ImageIds
	}
	return []int{}
}
func (c *Create) GetModel() string {
	return c.Model
}
func (c *Create) GetPrompt() string {
	return c.Prompt
}
func (c *Create) GetQuality() string {
	return ""
}
func (c *Create) GetSize() string {
	return c.Size
}
func (c *Create) GetSpeed() consts.TaskSpeed {
	return ""
}
func (c *Create) GetTaskType() string {
	if len(c.ImageIds) != 0 {
		return consts.TaskTypeEdit.String()
	}
	return consts.TaskTypeGenerate.String()
}
