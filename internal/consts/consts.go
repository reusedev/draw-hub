package consts

type ModelSupplier string

const (
	Geek ModelSupplier = "geek"
	Tuzi ModelSupplier = "tuzi"
	V3   ModelSupplier = "v3"
)

func (m ModelSupplier) String() string {
	return string(m)
}

func (m ModelSupplier) BaseURL() string {
	switch m {
	case Geek:
		return "https://geekai.co/api"
	case Tuzi:
		return "https://api.tu-zi.com"
	case V3:
		return "https://api.gpt.ge"
	default:
		return ""
	}
}

type Model string

const (
	GPT4oImage    Model = "gpt-4o-image"
	GPT4oImageVip Model = "gpt-4o-image-vip"
	GPTImage1     Model = "gpt-image-1"
)

func (m Model) String() string {
	return string(m)
}

type TaskSpeed string

const (
	SlowSpeed TaskSpeed = "slow"
	FastSpeed TaskSpeed = "fast"
)

func (s TaskSpeed) String() string {
	return string(s)
}

const (
	FourImagePrompt = "\n请一定返回 2 张单独的图片,且 2 张图片内容有些差异化，便于我挑选"
)

type TaskType string

const (
	TaskTypeEdit     TaskType = "edit"
	TaskTypeGenerate TaskType = "generate"
)

func (t TaskType) String() string {
	return string(t)
}

const (
	EventPoll = "poll"
)
