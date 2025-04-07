package chatcmd

import (
	"github.com/mslacken/kowalski/internal/app/chat"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Ask kowalski what to change",
	Long: `iStart a chat with Kowalski, you helpfull penguin.
He has access to knowledge bases and can access your files
for better answers.`,
	Run: func(cmd *cobra.Command, args []string) {
		modelname, _ := cmd.PersistentFlags().GetString("model")
		settings := ollamaconnector.OllamaChat()
		settings.Model = modelname
		chat.Chat(&settings)
	},
}

func init() {
	chatCmd.PersistentFlags().StringP("model", "m", ollamaconnector.DefaultModel, "model name")
}

func GetCommand() *cobra.Command {
	return chatCmd
}
