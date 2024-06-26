package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"owlllovo/ginessential/common"
	"path/filepath"
)

func EncodeImageToBase64(imagePath string) (string, error) {
	// 读取图片文件内容
	imageBytes, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	// 将图片内容编码为base64字符串
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	return base64Image, nil
}

type Config struct {
	MaxTokens int `json:"MaxTokens"`
}

func GetGPTComment(imageFilename, promptText string) (string, error) {
	log.Println("Running GetGPTComment for image:", imageFilename, "with prompt:", promptText)
	imagePath := filepath.Join("assets", "images", imageFilename)
	imageBase64, err := EncodeImageToBase64(imagePath)
	if err != nil {
		log.Printf("Error encoding image to base64: %v", err)
		return "", err
	}
	type ImageURL struct {
		URL string `json:"url"`
	}

	type Content struct {
		Type     string    `json:"type"`
		Text     string    `json:"text,omitempty"`
		ImageURL *ImageURL `json:"image_url,omitempty"`
	}

	type Message struct {
		Role    string    `json:"role"`
		Content []Content `json:"content"`
	}

	type Payload struct {
		Model     string    `json:"model"`
		Messages  []Message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}

	// 使用配置中的 MaxTokens
	fmt.Println("MaxTokens from config:", common.AppConfig.MaxTokens)

	data := Payload{
		Model: "gpt-4-vision-preview",
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: promptText,
					},
					{
						Type:     "image_url",
						ImageURL: &ImageURL{URL: "data:image/jpeg;base64," + imageBase64},
					},
				},
			},
		},
		MaxTokens: common.AppConfig.MaxTokens,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %d", resp.StatusCode)
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Response Body: %s", string(bodyBytes))
		return "", errors.New("API request failed")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	type ResponseChoice struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	type Response struct {
		Choices []ResponseChoice `json:"choices"`
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return "", err
	}

	if len(response.Choices) > 0 && len(response.Choices[0].Message.Content) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", errors.New("no AI comment received")
}

func GetVisualGLMComment(imageFilename, promptText string) (string, error) {
	log.Println("Calling VisualGLM for image:", imageFilename, "with description:", promptText)
	imagePath := filepath.Join("assets", "images", imageFilename)
	imageBase64, err := EncodeImageToBase64(imagePath)
	if err != nil {
		log.Printf("Error encoding image to base64: %v", err)
		return "", err
	}

	// Prepare the JSON payload
	type Payload struct {
		Image   string   `json:"image"`
		Text    string   `json:"text"`
		History []string `json:"history"`
	}

	data := Payload{
		Image:   imageBase64,
		Text:    promptText,
		History: []string{}, // This could be populated if there's contextual history needed
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		return "", err
	}

	// Set up the HTTP request
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Response Body: %s", string(bodyBytes))
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", err
	}

	// Assume the response is plain text or modify this part if it's structured JSON
	return string(body), nil
}
