package tools

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/webp"
)

func ConvertAndCompressPNGtoJPEG(srcData []byte, quality int) ([]byte, error) {
	if DetectImageType(srcData).String() != "png" {
		return nil, fmt.Errorf("not a PNG image")
	}
	// 解码PNG图像
	img, err := png.Decode(bytes.NewReader(srcData))
	if err != nil {
		return nil, err
	}
	// 设定JPEG压缩选项
	options := jpeg.Options{
		Quality: quality, // 质量范围从1到100，数值越高，质量越好，文件越大
	}
	ret := new(bytes.Buffer)
	// 将图像编码为JPEG并输出到文件
	err = jpeg.Encode(ret, img, &options)
	if err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}

func ConvertAndCompressToJPEG(srcData []byte, quality int) ([]byte, error) {
	imageType := DetectImageType(srcData)
	var img image.Image
	var err error
	switch imageType {
	case ImageTypePNG:
		img, err = png.Decode(bytes.NewReader(srcData))
	case ImageTypeJPEG:
		img, err = jpeg.Decode(bytes.NewReader(srcData))
	case ImageTypeWEBP:
		img, err = webp.Decode(bytes.NewReader(srcData))
	default:
		return nil, fmt.Errorf("unsupported image type: %s", imageType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	options := jpeg.Options{
		Quality: quality,
	}
	ret := new(bytes.Buffer)
	err = jpeg.Encode(ret, img, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}
	return ret.Bytes(), nil
}
