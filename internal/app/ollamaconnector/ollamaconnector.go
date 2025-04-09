package ollamaconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mslacken/kowalski/internal/pkg/templates"
)

var emblength *int

type Settings struct {
	Model           string
	EmbeddingModel  string
	OllamaURL       string
	EmbeddingLength int
}

func OllamaOffline() Settings {
	return Settings{
		Model:          "gemma3:4b",
		EmbeddingModel: "nomic-embed-text",
		OllamaURL:      "http://localhost:11434/api/",
	}
}

func Ollama() Settings {
	sett := OllamaOffline()
	if emblength == nil {
		length, _ := sett.GetEmbeddingSize()
		emblength = &length
	}
	sett.EmbeddingLength = *emblength
	return sett
}

type TaskRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Format  string         `json:"format"`
	Options map[string]any `json:"options"`
	Stream  bool           `json:"stream"`
	System  string         `json:"system"`
}

type GenerateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	Tool_Calls []any  `json:"tool_calls"`
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

type TaskResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

type ModelInfo struct {
	License       string         `json:"license,omitempty"`
	Modelfile     string         `json:"modelfile,omitempty"`
	Parameters    string         `json:"parameters,omitempty"`
	Template      string         `json:"template,omitempty"`
	System        string         `json:"system,omitempty"`
	ModelInfo     map[string]any `json:"model_info,omitempty"`
	ProjectorInfo map[string]any `json:"projector_info,omitempty"`
	ModifiedAt    time.Time      `json:"modified_at,omitempty"`
}

func (settings Settings) SendTask(msg string) (resp *TaskResponse, err error) {
	req := TaskRequest{
		Prompt: msg,
		System: templates.SystemPrompt,
		Model:  settings.Model,
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/generate"
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal message: %s", err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.Model, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.Model, err)
	}
	defer httpResp.Body.Close()
	ollamaResp := TaskResponse{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}

func (settings Settings) GetEmbeddings(emb []string) (*EmbeddingResponse, error) {
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/embed"
	req := EmbeddingRequest{
		Input: emb,
		Model: settings.EmbeddingModel,
	}
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	defer httpResp.Body.Close()
	var ollamaResp EmbeddingResponse
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	if err != nil {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		fmt.Printf("Body\n %s\n", string(bodyBytes))
	}
	return &ollamaResp, err
}

func (settings Settings) GetEmbeddingSize() (int, error) {
	info, err := settings.GetEmbeddingInfo()
	if err != nil {
		return 0, err
	}
	if modelArch, ok := info.ModelInfo["general.architecture"].(string); ok {
		// Follwing if clause doesn't work, I don't know why
		// if _, ok_l := info.ModelInfo["nomic-bert.embedding_length"].(int32); ok_l {
		return int(info.ModelInfo[modelArch+".embedding_length"].(float64)), nil
		// }
		// return 0, fmt.Errorf("couldn't determine embdding length of %s", modelArch)
	}
	return 0, fmt.Errorf("couldn't get model info")

}
func (settings Settings) GetEmbeddingInfo() (*ModelInfo, error) {
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/show"
	var req = struct {
		Model   string `json:"model,omitempty"`
		Verbose bool   `json:"verbos,omitempty"`
	}{
		Model:   settings.EmbeddingModel,
		Verbose: false,
	}
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s EmbeddingModel: %s Error: %v", URL, settings.EmbeddingModel, err)
	}
	defer httpResp.Body.Close()
	var info ModelInfo
	err = json.NewDecoder(httpResp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (settings Settings) GetReq(req GenerateReq) (str string, err error) {
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/generate"
	js, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.Model, err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return "", fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.Model, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.Model, err)
	}
	defer httpResp.Body.Close()
	bodyBytes, _ := io.ReadAll(httpResp.Body)
	return string(bodyBytes), nil
}
