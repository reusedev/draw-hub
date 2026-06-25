package gemini

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

func newFlashImageTestResponse(supplier, tokenDesc, model string) *FlashImageResponse {
	return &FlashImageResponse{
		BaseResponse: image.BaseResponse{
			Supplier:  supplier,
			TokenDesc: tokenDesc,
			Model:     model,
			URLs:      []string{},
		},
	}
}

func TestFlashImageParser_Parse(t *testing.T) {
	parser := NewFlashImageParser()

	// 测试用例1：标准 markdown 格式响应
	t.Run("标准markdown格式响应", func(t *testing.T) {
		markdownBody := `data: {"choices":[{"delta":{"content":"![图片](https://example.com/test-image.png)"}}]}

data: [DONE]`

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(markdownBody)),
			Request: &http.Request{
				URL:    &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := newFlashImageTestResponse("test", "test", "gemini-2.5-flash-image")

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if !response.Succeed() {
			t.Fatalf("解析应该成功，但失败了")
		}

		expectedURL := "https://example.com/test-image.png"
		if len(response.URLs) == 0 {
			t.Fatalf("应该提取到URL，但没有提取到")
		}

		if response.URLs[0] != expectedURL {
			t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
		}

		t.Logf("成功提取到URL: %s", response.URLs[0])
	})

	// 测试用例2：JSON代码块格式响应（带绘画emoji） - Gemini 特有的格式
	t.Run("JSON代码块格式响应_Gemini", func(t *testing.T) {
		// 模拟流式响应的数据块
		streamData := []string{
			`data: {"choices":[{"delta":{"content":"> 🖌️正在绘画\n\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"` + "```json" + `\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"{\"model\":\"google/gemini-2.5-flash-image-preview:\",\"prompt\":\"去掉图片中的文字,图片比例:2:3\",\"n\":1,\"size\":\"1024x1024\",\"response_format\":\"url\",\"aspect_ratio\":\"1:1\",\"image\":[\"https://s3.ffire.cc/cdn/20250904/Nqz5BP8wAcXXz5EqBJ2aR7_None\"]}\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"` + "```" + `\n\n"}}]}`,
			`data: [DONE]`,
		}

		streamBody := strings.Join(streamData, "\n") + "\n"

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(streamBody)),
			Request: &http.Request{
				URL:    &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := newFlashImageTestResponse("test", "test", "gemini-2.5-flash-image-preview")

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if !response.Succeed() {
			t.Fatalf("解析应该成功，但失败了。响应体: %s", response.RespBody)
		}

		expectedURL := "https://s3.ffire.cc/cdn/20250904/Nqz5BP8wAcXXz5EqBJ2aR7_None"
		if len(response.URLs) == 0 {
			t.Fatalf("应该提取到URL，但没有提取到。响应体: %s", response.RespBody)
		}

		if response.URLs[0] != expectedURL {
			t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
		}

		t.Logf("成功提取到Gemini JSON代码块中的URL: %s", response.URLs[0])
	})

	// 测试用例3：多图片JSON代码块格式响应
	t.Run("多图片JSON代码块格式响应_Gemini", func(t *testing.T) {
		streamData := []string{
			`data: {"choices":[{"delta":{"content":"> 🖌️正在生成多张图片\n\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"` + "```json" + `\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"{\"model\":\"gemini-test-model\",\"prompt\":\"测试多图片生成\",\"n\":2,\"image\":[\"https://example.com/image1.jpg\",\"https://example.com/image2.png\"]}\n"}}]}`,
			`data: {"choices":[{"delta":{"content":"` + "```" + `\n\n处理完成！"}}]}`,
			`data: [DONE]`,
		}

		streamBody := strings.Join(streamData, "\n") + "\n"

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(streamBody)),
			Request: &http.Request{
				URL:    &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := newFlashImageTestResponse("test", "test", "gemini-test-model")

		err := parser.Parse(resp, response)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if !response.Succeed() {
			t.Fatalf("解析应该成功，但失败了。响应体: %s", response.RespBody)
		}

		expectedURLs := []string{
			"https://example.com/image1.jpg",
			"https://example.com/image2.png",
		}

		if len(response.URLs) != len(expectedURLs) {
			t.Fatalf("期望URL数量: %d, 实际URL数量: %d", len(expectedURLs), len(response.URLs))
		}

		for i, expectedURL := range expectedURLs {
			if response.URLs[i] != expectedURL {
				t.Fatalf("期望URL[%d]: %s, 实际URL[%d]: %s", i, expectedURL, i, response.URLs[i])
			}
		}

		t.Logf("成功提取到多个Gemini图片URL: %v", response.URLs)
	})

	// 测试用例4：无效响应
	t.Run("无效响应", func(t *testing.T) {
		invalidBody := `data: {"choices":[{"delta":{"content":"无效内容，没有图片"}}]}

data: [DONE]`

		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(invalidBody)),
			Request: &http.Request{
				URL:    &url.URL{Path: "/v1/chat/completions"},
				Method: "POST",
			},
		}

		response := newFlashImageTestResponse("test", "test", "gemini-test")

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

// Gemini 集成测试
func TestFlashImageParser_Integration(t *testing.T) {
	parser := NewFlashImageParser()

	// 模拟您提供的实际 Gemini 响应格式
	streamData := []string{
		`data: {"choices":[{"delta":{"content":"> 🖌️正在绘画\n\n"}}]}`,
		`data: {"choices":[{"delta":{"content":"` + "```json" + `\n"}}]}`,
		`data: {"choices":[{"delta":{"content":"{\"model\":\"google/gemini-2.5-flash-image-preview:\",\"prompt\":\"去掉图片中的文字,图片比例:2:3\",\"n\":1,\"size\":\"1024x1024\",\"response_format\":\"url\",\"aspect_ratio\":\"1:1\",\"image\":[\"https://s3.ffire.cc/cdn/20250904/Nqz5BP8wAcXXz5EqBJ2aR7_None\"]}\n"}}]}`,
		`data: {"choices":[{"delta":{"content":"` + "```" + `\n\n"}}]}`,
		`data: [DONE]`,
	}

	streamBody := strings.Join(streamData, "\n") + "\n"

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(streamBody)),
		Request: &http.Request{
			URL:    &url.URL{Path: "/v1/chat/completions"},
			Method: "POST",
		},
	}

	response := newFlashImageTestResponse("gemini", "default", "gemini-2.5-flash-image-preview")

	err := parser.Parse(resp, response)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if !response.Succeed() {
		t.Fatalf("解析应该成功，但失败了。响应体: %s", response.RespBody)
	}

	expectedURL := "https://s3.ffire.cc/cdn/20250904/Nqz5BP8wAcXXz5EqBJ2aR7_None"
	if len(response.URLs) == 0 {
		t.Fatalf("应该提取到URL，但没有提取到")
	}

	if response.URLs[0] != expectedURL {
		t.Fatalf("期望URL: %s, 实际URL: %s", expectedURL, response.URLs[0])
	}

	t.Logf("✅ Gemini集成测试通过！成功提取到URL: %s", response.URLs[0])
}
