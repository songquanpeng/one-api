package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// ChatCompletionRequest defines the request structure
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message defines the message structure
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionChunk defines the chunk structure for streaming response
type ChatCompletionChunk struct {
	Choices []Choice `json:"choices"`
}

// Choice defines the choice structure
type Choice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
}

// Config defines the OpenAI API configuration
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Key  string `json:"key"`
}

func Chatting(config Config) {
	apiKey := config.Key
	messages := []Message{}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter your message: ")
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()
		messages = append(messages, Message{Role: "user", Content: userInput})

		request := ChatCompletionRequest{
			Model:    "ERNIE-3.5-8K",
			Messages: messages,
			Stream:   true,
		}

		requestBody, err := json.Marshal(request)
		if err != nil {
			log.Fatalf("Error encoding JSON request body: %v", err)
		}

		url := fmt.Sprintf("%s:%d/v1/chat/completions", config.Host, config.Port)
		req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Fatalf("Error creating HTTP request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error sending HTTP request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("Request failed, status code: %d, response content: %s", resp.StatusCode, string(body))
		}

		reader := bufio.NewReader(resp.Body)
		totalResponse := ""
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("Error reading response line: %v", err)
			}

			if len(line) > 6 && string(line[:6]) == "data: " {
				line = line[6:]
				if strings.Contains(string(line), "[DONE]") {
					break
				}

				var chunk ChatCompletionChunk
				err = json.Unmarshal(line, &chunk)
				if err != nil {
					log.Printf("Error parsing JSON chunk: %v, data: %s", err, string(line))
					continue
				}

				for _, choice := range chunk.Choices {
					fmt.Print(choice.Delta.Content)
					totalResponse += choice.Delta.Content
				}
			}
		}
		messages = append(messages, Message{Role: "assistant", Content: totalResponse})
		fmt.Println()
	}
}
