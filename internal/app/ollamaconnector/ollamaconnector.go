package ollamaconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}
type Ollama struct {
	Url string
	Request
}

func (ollama *Ollama) TalkToOllama(msg Message) (*Response, error) {
	ollama.Request.Messages = []Message{msg}
	js, err := json.Marshal(ollama)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", ollama.Url, ollama.Model, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, ollama.Url, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", ollama.Url, ollama.Model, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", ollama.Url, ollama.Model, err)
	}
	defer httpResp.Body.Close()
	ollamaResp := Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}
