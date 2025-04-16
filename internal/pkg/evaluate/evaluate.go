package evaluate

type Evaluation struct {
	// name of evluation, must be provided
	Name     string   `yaml:"name"`
	Prompt   string   `yaml:"prompt"`
	OS       string   `yaml:"OS,omitempty"`
	Context  string   `yaml:"context,omitempty"`
	Response string   `yaml:"response,omitempty"`
	Answers  []string `yaml:"answers,omitempty"`
	Files    []*File  `yaml:"files,omitempty"`
}

type File struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content,omitempty"`
}

type EvalutaionList struct {
	// uuid of evaluation after run
	Id          string        `yaml:"id,omitempty"`
	Version     string        `yaml:"version,omitempty"`
	LLM         string        `yaml:"llm,omitempty"`
	Embedding   string        `yaml:"embedding,omitempty"`
	Evaluations []*Evaluation `yaml:"evaluations"`
}
