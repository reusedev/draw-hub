package request

type TaskForm interface {
	GetGroupId() string
	GetImageIds() []int
	GetPrompt() string
	GetSpeed() string
	GetQuality() string
	GetSize() string
}

type SlowTask struct {
	GroupId  string `form:"group_id"`
	ImageId  int    `form:"image_id"`
	ImageIds []int  `form:"image_ids"`
	Prompt   string `form:"prompt"`
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
func (s *SlowTask) GetSpeed() string {
	return "slow"
}

type FastSpeed struct {
	GroupId  string `form:"group_id"`
	ImageId  int    `form:"image_id"`
	ImageIds []int  `form:"image_ids"`
	Prompt   string `form:"prompt"`
	Quality  string `form:"quality"`
	Size     string `form:"size"`
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
func (s *FastSpeed) GetSpeed() string {
	return "fast"
}
