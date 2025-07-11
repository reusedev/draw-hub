package gpt

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestImage4oParser_Parse(t *testing.T) {
	parser := &Image4oParser{}

	// 测试用例1：新的JSON格式响应（你提供的日志数据）
	t.Run("JSON格式响应", func(t *testing.T) {
		jsonBody := `{
			"id": "foaicmpl-d58228d0-91fa-456d-8769-54fa43e6a18f",
			"model": "sora_image",
			"object": "chat.completion",
			"created": 1752157668,
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "![图片](https://midjourney-plus.oss-us-west-1.aliyuncs.com/sora/b5e92b94-f83b-47e5-9b64-af64ddeee928.png)\n\n"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 466,
				"completion_tokens": 47,
				"total_tokens": 513
			}
		}`

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(jsonBody)),
			Request: &http.Request{
				URL: &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := &Image4oResponse{
			Supplier:  "v3",
			TokenDesc: "default",
			Model:     "gpt-4o-image-vip",
			Duration:  111890 * time.Microsecond,
		}

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if !response.Succeed() {
			t.Fatalf("解析应该成功，但失败了")
		}

		expectedURL := "https://midjourney-plus.oss-us-west-1.aliyuncs.com/sora/b5e92b94-f83b-47e5-9b64-af64ddeee928.png"
		if len(response.URLs) == 0 {
			t.Fatalf("应该提取到URL，但没有提取到")
		}

		if response.URLs[0] != expectedURL {
			t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
		}

		t.Logf("成功提取到URL: %s", response.URLs[0])
	})

	// 测试用例2：原来的格式响应（向后兼容性测试）
	t.Run("原格式响应", func(t *testing.T) {
		oldBody := `(https://example.com/image.jpg)\n\n[点击下载]`

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(oldBody)),
			Request: &http.Request{
				URL: &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := &Image4oResponse{
			Supplier:  "test",
			TokenDesc: "test",
			Model:     "test",
			Duration:  1000 * time.Microsecond,
		}

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if !response.Succeed() {
			t.Fatalf("解析应该成功，但失败了")
		}

		expectedURL := "https://example.com/image.jpg"
		if len(response.URLs) == 0 {
			t.Fatalf("应该提取到URL，但没有提取到")
		}

		if response.URLs[0] != expectedURL {
			t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
		}

		t.Logf("成功提取到URL: %s", response.URLs[0])
	})

	// 测试用例3：无效响应
	t.Run("无效响应", func(t *testing.T) {
		invalidBody := `{"invalid": "json"}`

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(invalidBody)),
			Request: &http.Request{
				URL: &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := &Image4oResponse{
			Supplier:  "test",
			TokenDesc: "test",
			Model:     "test",
			Duration:  1000 * time.Microsecond,
		}

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if response.Succeed() {
			t.Fatalf("无效响应应该失败，但成功了")
		}

		t.Logf("正确处理了无效响应")
	})
}

// 简单的集成测试
func TestImage4oParser_Integration(t *testing.T) {
	parser := &Image4oParser{}

	// 模拟你提供的实际日志数据
	jsonBody := `{
		"id": "foaicmpl-d58228d0-91fa-456d-8769-54fa43e6a18f",
		"model": "sora_image",
		"object": "chat.completion",
		"created": 1752157668,
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "![图片](https://midjourney-plus.oss-us-west-1.aliyuncs.com/sora/b5e92b94-f83b-47e5-9b64-af64ddeee928.png)\n\n"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 466,
			"completion_tokens": 47,
			"total_tokens": 513,
			"prompt_tokens_details": {
				"cached_tokens": 0,
				"text_tokens": 0,
				"audio_tokens": 0,
				"image_tokens": 0
			},
			"completion_tokens_details": {
				"text_tokens": 0,
				"audio_tokens": 0,
				"reasoning_tokens": 0
			},
			"input_tokens": 0,
			"output_tokens": 0,
			"input_tokens_details": null
		}
	}`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
		Request: &http.Request{
			URL: &url.URL{Path: "/v1/chat/completions"},
			Method: "POST",
		},
	}

	response := &Image4oResponse{
		Supplier:  "v3",
		TokenDesc: "default",
		Model:     "gpt-4o-image-vip",
		Duration:  111890 * time.Microsecond,
	}

	err := parser.Parse(resp, response)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if !response.Succeed() {
		t.Fatalf("解析应该成功，但失败了。响应体: %s", response.RespBody)
	}

	expectedURL := "https://midjourney-plus.oss-us-west-1.aliyuncs.com/sora/b5e92b94-f83b-47e5-9b64-af64ddeee928.png"
	if len(response.URLs) == 0 {
		t.Fatalf("应该提取到URL，但没有提取到")
	}

	if response.URLs[0] != expectedURL {
		t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
	}

	t.Logf("✅ 集成测试通过！成功提取到URL: %s", response.URLs[0])
}
