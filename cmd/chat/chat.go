package chatcmd

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/app/chat"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
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
		chat.Chat(&ollamaconnector.Ollamasettings)
	},
}

// send jus a simple request from the command line, is hidden
// as intended for testing and debugging
var reqCmd = &cobra.Command{
	Use:   "request",
	Short: "send request from commandline",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		if err != nil {
			return err
		}
		prompt, err := db.GetContext(args[0], []string{}, ollamaconnector.Ollamasettings.GetContextSize())
		if err != nil {
			return err
		}
		log.Infof("Prompt: %s", prompt)
		ch := make(chan *ollamaconnector.TaskResponse)
		respStr := []string{}
		go ollamaconnector.Ollamasettings.SendTaskStream(prompt, ch)
		for resp := range ch {
			respStr = append(respStr, resp.Response)
			log.Debug(resp.Response)
		}
		log.Infof("Kowalski: %s", strings.Join(respStr, ``))
		return nil
	},
	Args:   cobra.MinimumNArgs(1),
	Hidden: true,
}

func init() {
	chatCmd.AddCommand(reqCmd)
}

func GetCommand() *cobra.Command {
	return chatCmd
}
