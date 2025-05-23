package image

type Request struct {
	Speed    string `json:"speed"`
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt"`
	Quality  string `json:"quality"`
	Size     string `json:"size"`
}

func SlowSpeed(request Request) {

}

func FastSpeed(request Request) {

}
