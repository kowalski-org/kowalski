package evaluatecmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/evaluate"
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
		sett := ollamaconnector.Ollama()
		evaluationList := evaluate.EvalutaionList{
			Id:        id.String(),
			Version:   version.Commit,
			LLM:       sett.LLM,
			Embedding: sett.EmbeddingModel,
		}
		log.Infof("starting evaluation with id: %s", id.String())
		log.Infof("LLM: %s embedding: %s", sett.LLM, sett.EmbeddingModel)
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
		for i, eval := range evaluationList.Evaluations {
			log.Infof("on evaluation '%s'", eval.Name)
			prompt, err := db.GetContext(eval.Prompt, []string{}, sett.ContextLength)
			if err != nil {
				return err
			}
			resp, err := sett.SendTask(prompt)
			evaluationList.Evaluations[i].Response = resp.Response
			if context {
				evaluationList.Evaluations[i].Context = prompt
			}
		}
		yml, err := yaml.Marshal(evaluationList)
		if err != nil {
			return err
		}
		err = os.WriteFile(id.String()+".yaml", yml, 0644)
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
