package image

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"regexp"
	"strings"
)

type URLParseStrategy interface {
	ExtractURLs(body []byte) ([]string, error)
}

type B64ParseStrategy interface {
	ExtractB64s(body []byte) ([]string, error)
}

type MarkdownURLStrategy struct{}

func (m *MarkdownURLStrategy) ExtractURLs(body []byte) ([]string, error) {
	var urls []string
	var content string

	// 首先尝试解析JSON格式的聊天完成响应
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := jsoniter.Unmarshal(body, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	} else {
		// 如果不是JSON格式，直接使用原始body作为内容
		content = string(body)
	}

	// 尝试提取markdown格式的图片链接
	markdownReg := `!\[.*?\]\((https?://[^)]+)\)`
	pattern, _ := regexp.Compile(markdownReg)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			url := match[1]
			url = strings.ReplaceAll(url, "\\u0026", "&")
			urls = append(urls, url)
		}
	}

	// 尝试解析JSON代码块中的图片链接（无论是否找到markdown图片）
	jsonBlockReg := "```json\\s*\\n([\\s\\S]*?)\\n```"
	jsonPattern, _ := regexp.Compile(jsonBlockReg)
	jsonMatches := jsonPattern.FindAllStringSubmatch(content, -1)
	for _, jsonMatch := range jsonMatches {
		if len(jsonMatch) >= 2 {
			var jsonData struct {
				Image []string `json:"image"`
			}
			if err := jsoniter.Unmarshal([]byte(jsonMatch[1]), &jsonData); err == nil {
				for _, imageURL := range jsonData.Image {
					if imageURL != "" {
						imageURL = strings.ReplaceAll(imageURL, "\\u0026", "&")
						urls = append(urls, imageURL)
					}
				}
			}
		}
	}

	// 如果仍然没有找到URL，尝试原来的正则表达式作为后备
	if len(urls) == 0 {
		reg := `(https?[^)]+)\)`
		pattern, _ := regexp.Compile(reg)
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				url := match[1]
				url = strings.ReplaceAll(url, "\\u0026", "&")
				urls = append(urls, url)
			}
		}
	}

	return urls, nil
}

type OpenAIURLStrategy struct{}

func (o *OpenAIURLStrategy) ExtractURLs(body []byte) ([]string, error) {
	var urls []string
	var s struct {
		Data []struct {
			URL           string `json:"url,omitempty"`
			B64JSON       string `json:"b64_json,omitempty"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		} `json:"data"`
	}
	err := jsoniter.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	for _, v := range s.Data {
		if v.URL != "" {
			urls = append(urls, v.URL)
		}
	}
	return urls, nil
}

type GenericB64Strategy struct{}

func (m *GenericB64Strategy) ExtractB64s(body []byte) ([]string, error) {
	var b64s []string
	input := string(body)
	prefix := "base64,"
	index := strings.Index(input, prefix)
	if index == -1 {
		return nil, fmt.Errorf("base64 prefix not found")
	}
	var b64 string
	if strings.HasSuffix(input, ")") {
		b64 = input[index+len(prefix) : len(input)-1]
	} else {
		b64 = input[index+len(prefix):]
	}
	b64s = append(b64s, b64)
	return b64s, nil
}
