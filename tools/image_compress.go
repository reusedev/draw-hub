package tools

import (
	"bytes"
	"fmt"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
)

func ConvertToPNG(srcData []byte) ([]byte, error) {
	imageType := DetectImageType(srcData)
	// 如果已经是 PNG，直接返回原数据
	if imageType == ImageTypePNG {
		return srcData, nil
	}

	var img image.Image
	var err error
	switch imageType {
	case ImageTypeJPEG:
		img, err = jpeg.Decode(bytes.NewReader(srcData))
	case ImageTypeGIF:
		img, err = gif.Decode(bytes.NewReader(srcData))
	case ImageTypeWEBP:
		img, err = webp.Decode(bytes.NewReader(srcData))
	//case ImageTypeHEIC:
	//	img, err = decodeHEIC(bytes.NewReader(srcData))
	default:
		return nil, fmt.Errorf("unsupported image type: %s", imageType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	ret := new(bytes.Buffer)
	err = png.Encode(ret, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}
	return ret.Bytes(), nil
}

//func decodeHEIC(reader io.Reader) (image.Image, error) {
//	// 1️⃣ 读取文件
//	data, err := io.ReadAll(reader)
//	if err != nil {
//		return nil, err
//	}
//
//	// 2️⃣ 创建 context
//	ctx, err := heif.NewContext()
//	if err != nil {
//		return nil, err
//	}
//
//	// 3️⃣ 读取 HEIC 数据
//	err = ctx.ReadFromMemory(data)
//	if err != nil {
//		return nil, err
//	}
//
//	// 4️⃣ 获取主图像 handle
//	handle, err := ctx.GetPrimaryImageHandle()
//	if err != nil {
//		return nil, err
//	}
//
//	// 5️⃣ 解码成图像
//	img, err := handle.DecodeImage(heif.ColorspaceRGB, heif.ChromaInterleavedRGB, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	// 6️⃣ 转成 Go image.Image
//	return img.GetImage()
//}

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
	//case ImageTypeHEIC:
	//	img, err = decodeHEIC(bytes.NewReader(srcData))
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
