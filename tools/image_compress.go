package tools

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
)

func ConvertAndCompressPNGtoJPEG(srcData []byte, quality int) ([]byte, error) {
	if DetectImageType(srcData) != "png" {
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
