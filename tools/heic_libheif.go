//go:build cgo

package tools

/*
#cgo pkg-config: libheif
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <libheif/heif.h>
*/
import "C"

import (
	"fmt"
	"image"
	"io"
	"unsafe"
)

func decodeHEIC(reader io.Reader) (image.Image, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty HEIC image")
	}

	ctx := C.heif_context_alloc()
	if ctx == nil {
		return nil, fmt.Errorf("allocate HEIC context")
	}
	defer C.heif_context_free(ctx)

	heifErr := C.heif_context_read_from_memory_without_copy(ctx, unsafe.Pointer(&data[0]), C.size_t(len(data)), nil)
	if heifErr.code != C.heif_error_Ok {
		return nil, heifError("read HEIC data", heifErr)
	}

	var handle *C.struct_heif_image_handle
	heifErr = C.heif_context_get_primary_image_handle(ctx, &handle)
	if heifErr.code != C.heif_error_Ok {
		return nil, heifError("get primary HEIC image", heifErr)
	}
	defer C.heif_image_handle_release(handle)

	var decoded *C.struct_heif_image
	heifErr = C.heif_decode_image(handle, &decoded, C.heif_colorspace_RGB, C.heif_chroma_interleaved_RGBA, nil)
	if heifErr.code != C.heif_error_Ok {
		return nil, heifError("decode HEIC image", heifErr)
	}
	defer C.heif_image_release(decoded)

	width := int(C.heif_image_get_width(decoded, C.heif_channel_interleaved))
	height := int(C.heif_image_get_height(decoded, C.heif_channel_interleaved))
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid HEIC image dimensions: %dx%d", width, height)
	}

	var stride C.int
	plane := C.heif_image_get_plane_readonly(decoded, C.heif_channel_interleaved, &stride)
	if plane == nil {
		return nil, fmt.Errorf("read decoded HEIC image plane")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rowBytes := width * 4
	for y := 0; y < height; y++ {
		src := unsafe.Pointer(uintptr(unsafe.Pointer(plane)) + uintptr(y)*uintptr(stride))
		dstStart := y * img.Stride
		copy(img.Pix[dstStart:dstStart+rowBytes], C.GoBytes(src, C.int(rowBytes)))
	}

	return img, nil
}

func heifError(action string, err C.struct_heif_error) error {
	message := "unknown error"
	if err.message != nil {
		message = C.GoString(err.message)
	}
	return fmt.Errorf("%s: %s", action, message)
}
