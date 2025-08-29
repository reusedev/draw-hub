package request

import "github.com/reusedev/draw-hub/internal/consts"

type TaskForm interface {
	GetImageOrigin() string
	GetGroupId() string
	GetImageIds() []int
	GetPrompt() string
	GetSpeed() consts.TaskSpeed
	GetQuality() string
	GetSize() string
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

type Generate struct {
	GroupId string `form:"group_id"`
	Prompt  string `form:"prompt"`
}

type Create struct {
	ImageType string `form:"image_type"`
	GroupId   string `form:"group_id"`
	ImageIds  []int  `form:"image_ids"`
	Prompt    string `form:"prompt"`
}
