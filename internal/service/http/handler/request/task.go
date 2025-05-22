package request

type SlowTask struct {
	GroupId string `form:"group_id"`
	ImageId int    `form:"image_id"`
	Prompt  string `form:"prompt"`
}

type FastSpeed struct {
	GroupId string `form:"group_id"`
	ImageId int    `form:"image_id"`
	Prompt  string `form:"prompt"`
	Quality string `form:"quality"`
	Size    string `form:"size"`
}
