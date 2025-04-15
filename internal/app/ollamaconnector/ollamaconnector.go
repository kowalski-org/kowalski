package ollamaconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// context and embeeing lenths shouldn't change so creat global
// var for that
var emblength *int
var contlength *int

type Settings struct {
	LLM             string
	EmbeddingModel  string
	OllamaURL       string
	EmbeddingLength int
	ContextLength   int
}

func OllamaOffline() Settings {
	return Settings{
		LLM:            viper.GetString("llm"),
		EmbeddingModel: viper.GetString("embedding"),
		OllamaURL:      viper.GetString("URL"),
	}
}

func Ollama() Settings {
	sett := OllamaOffline()
	if emblength == nil {
		emblength = new(int)
		*emblength, _ = sett.GetEmbeddingSize()
	}
	sett.EmbeddingLength = *emblength
	if contlength == nil {
		contlength = new(int)
		*contlength, _ = sett.GetContextSize()
	}
	sett.ContextLength = *contlength
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
		Prompt:  msg,
		Model:   settings.LLM,
		Options: map[string]any{"Temperature": 0},
		Stream:  false,
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/generate"
	js, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal message: %s", err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.LLM, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("URL: %s Model: %s Error: %v", URL, settings.LLM, err)
	}
	defer httpResp.Body.Close()
	ollamaResp := TaskResponse{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}
func (settings Settings) SendTaskStream(msg string, resp chan *TaskResponse) (err error) {
	req := TaskRequest{
		Prompt: msg,
		// System: templates.SystemPrompt,
		Model:   settings.LLM,
		Options: map[string]any{"Temperature": 0},
		Stream:  true,
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/generate"
	js, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("couldn't marshal message: %s", err)
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	if err != nil {
		return fmt.Errorf("Error when creating Request: URL: %s Model: %s Error: %v", URL, settings.LLM, err)
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Error during request retrival: URL: %s Model: %s Error: %v", URL, settings.LLM, err)
	}
	dec := json.NewDecoder(httpResp.Body)
	for {
		ollamaResp := TaskResponse{}
		err = dec.Decode(&ollamaResp)
		if err != nil {
			break
		}
		resp <- &ollamaResp
	}
	close(resp)
	return
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
		return nil, fmt.Errorf("couldn't decode respones: %s", err)
	}
	return &ollamaResp, err
}

/*
Get the embeddig size
*/
func (settings Settings) GetEmbeddingSize() (int, error) {
	info, err := settings.GetModelInfo(settings.EmbeddingModel)
	if err != nil {
		return 0, err
	}
	if modelArch, ok := info.ModelInfo["general.architecture"].(string); ok {
		return int(info.ModelInfo[modelArch+".embedding_length"].(float64)), nil
	} else {
		log.Warnf("couldn't get embedding size for %s", modelArch)
	}
	return 0, fmt.Errorf("couldn't get model info")
}

/*
Get the context size
*/
func (settings Settings) GetContextSize() (int, error) {
	info, err := settings.GetModelInfo(settings.LLM)
	if err != nil {
		return 0, err
	}
	if modelArch, ok := info.ModelInfo["general.architecture"].(string); ok {
		return int(info.ModelInfo[modelArch+".context_length"].(float64)), nil
	} else {
		log.Warnf("couldn't get context size for %s", modelArch)
	}
	return 0, fmt.Errorf("couldn't get model info")
}

/*
Get the basic information of the model via the REST API from ollma
*/
func (settings Settings) GetModelInfo(name string) (*ModelInfo, error) {
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/show"
	var req = struct {
		Model   string `json:"model,omitempty"`
		Verbose bool   `json:"verbos,omitempty"`
	}{
		Model:   name,
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
