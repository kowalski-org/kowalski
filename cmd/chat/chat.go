package chatcmd

import (
	"fmt"
	"strings"

	"github.com/mslacken/kowalski/internal/app/chat"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
	"github.com/mslacken/kowalski/internal/pkg/database"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Ask kowalski what to change",
	Long: `Start a chat with Kowalski, you helpfull penguin.
He has access to knowledge bases and can access your files
for better answers.`,
	Run: func(cmd *cobra.Command, args []string) {
		// modelname, _ := cmd.PersistentFlags().GetString("model")
		settings := ollamaconnector.Ollama()
		// settings.Model = modelname
		chat.Chat(&settings)
	},
}

var reqCmd = &cobra.Command{
	Use:   "req",
	Short: "send request from commandline",
	RunE: func(cmd *cobra.Command, args []string) error {
		context, err := database.GetContext(args[0], []string{})
		if err != nil {
			return err
		}
		prompt := strings.Join([]string{context, args[0]}, "\n")
		fmt.Println("Prompt:", prompt)
		sett := ollamaconnector.Ollama()
		ch := make(chan *ollamaconnector.TaskResponse)
		go sett.SendTaskStream(prompt, ch)
		for resp := range ch {
			fmt.Printf("%s", resp.Response)
		}
		fmt.Println()
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	// chatCmd.PersistentFlags().StringP("model", "m", ollamaconnector.Ollama().EmbeddingModel, "model name")
	chatCmd.AddCommand(reqCmd)
}

func GetCommand() *cobra.Command {
	return chatCmd
}
