package cmd

import (
	"github.com/mslacken/kowalski/internal/app/chat"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		modelname, _ := cmd.PersistentFlags().GetString("model")
		llm := ollamaconnector.Ollama{
			Url: "http://localhost:11434/api/chat",
			Request: ollamaconnector.Request{
				Model: modelname,
			},
		}
		chat.Chat(&llm)
	},
}

func init() {
	chatCmd.PersistentFlags().StringP("model", "m", "llama3.1", "model name")
	rootCmd.AddCommand(chatCmd)
}
