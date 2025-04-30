package ollamaconnector

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type Settings struct {
	LLM            string
	EmbeddingModel string
	OllamaURL      string
	embeddingSize  int
	contextSize    int
	info           ModelInfo
}

var Ollamasettings Settings

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
	isSet         bool
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
	settings.PullModel(settings.LLM)
	req := TaskRequest{
		Prompt:  msg,
		Model:   settings.LLM,
		Options: map[string]any{"temperature": 0},
		Stream:  false,
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/generate"
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
	settings.PullModel(settings.LLM)
	req := TaskRequest{
		Prompt: msg,
		// System: templates.SystemPrompt,
		Model:   settings.LLM,
		Options: map[string]any{"temperature": 0},
		Stream:  true,
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/generate"
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
	settings.PullModel(settings.EmbeddingModel)
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/embed"
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
func (settings *Settings) GetEmbeddingSize() int {
	if settings.embeddingSize != 0 {
		return settings.embeddingSize
	}
	info, err := settings.GetModelInfo(settings.EmbeddingModel)
	if err != nil {
		log.Warnf("couldn't get embedding size: %s", err)
		return 0
	}
	if modelArch, ok := info.ModelInfo["general.architecture"].(string); ok {
		settings.embeddingSize = int(info.ModelInfo[modelArch+".embedding_length"].(float64))
		return settings.embeddingSize
	} else {
		log.Warnf("couldn't get embedding size for %s", modelArch)
	}
	return 0
}

/*
Get the context size
*/
func (settings *Settings) GetContextSize() int {
	if settings.contextSize != 0 {
		return settings.contextSize
	}
	info, err := settings.GetModelInfo(settings.LLM)
	if err != nil {
		log.Warnf("couldn't get context size: %s", err)
		return 0
	}
	if modelArch, ok := info.ModelInfo["general.architecture"].(string); ok {
		settings.contextSize = int(info.ModelInfo[modelArch+".context_length"].(float64))
		return settings.contextSize
	} else {
		log.Warnf("couldn't get context size for %s", modelArch)
	}
	return 0
}

/*
Get the basic information of the model via the REST API from ollma
*/
func (settings Settings) GetModelInfo(name string) (*ModelInfo, error) {
	if settings.info.isSet {
		return &settings.info, nil
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/show"
	var req = struct {
		Model   string `json:"model,omitempty"`
		Verbose bool   `json:"verbose,omitempty"`
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
	err = json.NewDecoder(httpResp.Body).Decode(&settings.info)
	if err != nil {
		return nil, err
	}
	settings.info.isSet = true
	return &settings.info, nil
}

// check for model on the ollam instance
func (settings *Settings) FindModel(name string) (found bool, err error) {
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/tags"
	modelResp := struct {
		Models []map[string]any `json:"Models"`
	}{}
	httpGet, err := http.Get(URL)
	if err != nil {
		return false, err
	}
	if httpGet.StatusCode != http.StatusOK {
		return false, errors.New("couldn't list models of ollana")
	}
	err = json.NewDecoder(httpGet.Body).Decode(&modelResp)
	if err != nil {
		return false, errors.New("couldn't parse models from server")
	}
	for _, it := range modelResp.Models {
		ok := it["name"] == name
		if ok {
			return ok, nil
		}
	}
	return false, nil
}

// pull model if not present
func (settings *Settings) PullModel(name string) (err error) {
	found, err := settings.FindModel(name)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	URL := strings.TrimSuffix(settings.OllamaURL, "/") + "/api/pull"
	var req = struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream,omitempty"`
	}{
		Model:  name,
		Stream: false,
	}
	js, _ := json.Marshal(req)
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(js))
	httpResp, err := client.Do(httpReq)
	if httpResp.StatusCode != http.StatusOK {
		return errors.New("couldn't pull modell")
	}
	reader := bufio.NewReader(httpResp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		log.Debug(string(line))
	}
	return nil
}
