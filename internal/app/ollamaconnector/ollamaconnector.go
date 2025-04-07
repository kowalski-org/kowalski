package ollamaconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var DefaultEmbeddingModel string
var DefaultModel string
var DefaultUrlChat string
var DefaultUrlEmbedding string
var DefaultEmbeddingDim int

func init() {
	DefaultEmbeddingModel = "mxbai-embed-large"
	DefaultModel = "llama3.1"
	DefaultUrlChat = "http://localhost:11434/api/chat"
	DefaultUrlEmbedding = "http://localhost:11434/api/embed"
	DefaultEmbeddingDim = 1024
}

type OllamaSettings struct {
	Model          string
	EmbeddingModel string
	UrlChat        string
	UrlEmbedding   string
}

func OllamaChat() OllamaSettings {
	return OllamaSettings{
		Model:          DefaultModel,
		EmbeddingModel: DefaultEmbeddingModel,
		UrlChat:        DefaultUrlChat,
		UrlEmbedding:   DefaultUrlEmbedding,
	}
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
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
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

func (settings OllamaSettings) TalkToOllama(msg []ChatMessage) (*ChatResponse, error) {
	req := ChatRequest{
		Messages: msg,
		Model:    settings.Model,
	}
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", settings.UrlChat, settings.Model, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, settings.UrlChat, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", settings.UrlChat, settings.Model, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", settings.UrlChat, settings.Model, err)
	}
	defer httpResp.Body.Close()
	var ollamaResp ChatResponse
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}

func (settings OllamaSettings) GetEmbeddings(emb []string) (*EmbeddingResponse, error) {
	req := EmbeddingRequest{
		Input: emb,
		Model: settings.EmbeddingModel,
	}
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", settings.UrlEmbedding, settings.EmbeddingModel, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, settings.UrlEmbedding, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", settings.UrlEmbedding, settings.EmbeddingModel, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", settings.UrlEmbedding, settings.EmbeddingModel, err)
	}
	defer httpResp.Body.Close()
	var ollamaResp EmbeddingResponse
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}
