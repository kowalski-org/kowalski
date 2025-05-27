package evaluatecmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/evaluate"
	"github.com/openSUSE/kowalski/internal/pkg/file"
	"github.com/openSUSE/kowalski/internal/pkg/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var runEvaluate = &cobra.Command{
	Use:   "evaluate evaluatefile1.yaml evaluatefile2.yaml ...",
	Short: "Run evaluates given by files.",
	Long: `Runs the evaluates given by the files and writes output to
new files with .$ID, where $ID is as random UUID.
So the input file foo.yaml will create an output like foo.yaml.abcd1234.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := cmd.Flags().GetBool("context")
		if err != nil {
			return err
		}
		id := uuid.New()
		evaluationList := evaluate.EvalutaionList{
			Id:        id.String(),
			Version:   version.Version,
			LLM:       ollamaconnector.Ollamasettings.LLM,
			Embedding: ollamaconnector.Ollamasettings.EmbeddingModel,
		}
		log.Infof("starting evaluation with id: %s", id.String())
		log.Infof("LLM: %s embedding: %s", evaluationList.LLM, evaluationList.Embedding)
		for _, fileName := range args {
			file, err := os.ReadFile(fileName)
			if err != nil {
				if os.IsNotExist(err) {
					log.Warnf("file %s doesn't exist", fileName)
					continue
				}
				return err
			}
			fList := evaluate.EvalutaionList{}
			err = yaml.Unmarshal(file, &fList)
			if err != nil {
				fevaluation := evaluate.Evaluation{}
				err = yaml.Unmarshal(file, &fevaluation)
				if err != nil {
					return err
				}
				evaluationList.Evaluations = append(evaluationList.Evaluations, &fevaluation)
			} else {
				evaluationList.Evaluations = append(evaluationList.Evaluations, fList.Evaluations...)
			}
		}
		db, err := database.New()
		if err != nil {
			return err
		}
		resultList := []evaluate.EvlatuationResult{}
		for _, eval := range evaluationList.Evaluations {
			log.Infof("on evaluation '%s'", eval.Name)
			log.Infof("prompt: %s", eval.Prompt)
			mock := file.Mock{
				Content: map[string]string{"foo": "baar"},
			}
			prompt, err := db.GetContext(eval.Prompt, []string{}, mock, ollamaconnector.Ollamasettings.GetContextSize())
			if err != nil {
				return err
			}
			resp, err := ollamaconnector.Ollamasettings.SendTask(prompt)
			result := evaluate.EvlatuationResult{
				Response:           resp.Response,
				TotalDuration:      resp.TotalDuration,
				LoadDuration:       resp.LoadDuration,
				PromptEvalCount:    resp.PromptEvalCount,
				PromptEvalDuration: resp.PromptEvalDuration,
				EvalCount:          resp.EvalCount,
				EvalDuration:       resp.EvalDuration,
				Evaluation:         *eval,
			}
			if context {
				result.Context = prompt
			}
			log.Infof("TotalDuration %d, LoadDuration %d, PromptEvalCount  %d, PromptEvalDuration %d, EvalCount %d, EvalDuration %d", resp.TotalDuration, resp.LoadDuration, resp.PromptEvalCount, resp.PromptEvalDuration, resp.EvalCount, resp.EvalDuration)
			log.Infof("response: %s", result.Response)
			resultList = append(resultList, result)
		}
		yml, err := yaml.Marshal(resultList)
		if err != nil {
			return err
		}
		err = os.WriteFile("eval-"+id.String()+".yaml", yml, 0644)
		return err
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	runEvaluate.Flags().Bool("context", false, "include context in output")
}

func GetCommand() *cobra.Command {
	return runEvaluate
}
