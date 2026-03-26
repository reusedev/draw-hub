package dao

import (
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/model"
)

func ImageById(id int) (model.Image, error) {
	var image model.Image
	err := mysql.DB.Model(&model.Image{}).Where("id = ?", id).First(&image).Error
	if err != nil {
		return model.Image{}, err
	}
	return image, nil
}

func InputImageById(id int) (model.InputImage, error) {
	var inputImage model.InputImage
	err := mysql.DB.Model(&model.InputImage{}).Where("id = ?", id).First(&inputImage).Error
	if err != nil {
		return model.InputImage{}, err
	}
	return inputImage, nil
}

func OutputImageById(id int) (model.OutputImage, error) {
	var outputImage model.OutputImage
	err := mysql.DB.Model(&model.OutputImage{}).Where("id = ?", id).First(&outputImage).Error
	if err != nil {
		return model.OutputImage{}, err
	}
	return outputImage, nil
}
