package consts

const (
	GeekBaseURL = "https://geekai.co/api"
	TuziBaseURL = "https://api.tu-zi.com"
	V3BaseUrl   = "https://api.gpt.ge"
)

type ModelSupplier string

const (
	Geek ModelSupplier = "geek"
	Tuzi ModelSupplier = "tuzi"
	V3   ModelSupplier = "v3"
)

func (m ModelSupplier) String() string {
	return string(m)
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
