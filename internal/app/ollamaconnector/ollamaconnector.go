package ollamaconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var EmbeddingModel string
var Model string
var Url string

func init() {
	EmbeddingModel = "all-minilm"
	Model = "llama3.1"
	Url = 
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Model           string    `json:"model"`
	Embeddings      []float64 `json:"embeddings"`
	TotalDuration   int64     `json:"total_duration"`
	LoadDuration    int       `json:"load_duration"`
	PromptEvalCount int       `json:"prompt_eval_count"`
}

type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration"`
	LoadDuration       int         `json:"load_duration"`
	PromptEvalCount    int         `json:"prompt_eval_count"`
	PromptEvalDuration int         `json:"prompt_eval_duration"`
	EvalCount          int         `json:"eval_count"`
	EvalDuration       int64       `json:"eval_duration"`
}
<<<<<<< HEAD
type OllamaChat struct {
	Url string
	ChatRequest
}

type OllamaEmbedding struct {
	Url string
	EmbeddingRequest
}

func (ollama *OllamaChat) TalkToOllama(msg ChatMessage) (*ChatResponse, error) {
	ollama.ChatRequest.Messages = []ChatMessage{msg}
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
	var ollamaResp ChatResponse
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}

func (ollama *OllamaEmbedding) GetEmbeddings(emb []string) (*EmbeddingResponse, error) {
	ollama.EmbeddingRequest.Input = emb
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
	var ollamaResp EmbeddingResponse
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}
