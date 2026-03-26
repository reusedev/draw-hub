package tools

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestImageCompress(t *testing.T) {
	u := "https://googlecdn.datas.systems/storage/response_images/287/2025/12/30/1767059034661264780_4275.png"
	resp, err := http.Get(u)
	if err != nil {
		t.Error(err)
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(b))
	imageType := DetectImageType(b)
	t.Log(imageType)
	if imageType != ImageTypeJPEG {
		t.Log("image type error")
		return
	}
	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
		return
	}
	buffer := &bytes.Buffer{}
	err = png.Encode(buffer, img)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(b), len(buffer.Bytes()))
	f, err := os.Create("test.png")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = f.Write(buffer.Bytes())
	if err != nil {
		t.Error(err)
	}
}
