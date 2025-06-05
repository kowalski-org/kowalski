package chatcmd

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/app/chat"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/file"
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
		locationStr, _ := cmd.Flags().GetString("location")
		location := file.Local{
			Chroot: locationStr,
		}
		chat.Chat(&ollamaconnector.Ollamasettings, location)
	},
}

// send just a simple request from the command line, is hidden
// as intended for testing and debugging
var reqCmd = &cobra.Command{
	Use:   "request",
	Short: "send request from commandline",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		if err != nil {
			return err
		}
		locationStr, _ := cmd.Flags().GetString("location")
		location := file.Local{
			Chroot: locationStr,
		}
		prompt, err := db.GetContext(args[0], []string{}, location, ollamaconnector.Ollamasettings.GetContextSize())
		if err != nil {
			return err
		}
		log.Infof("Prompt: %s", prompt)
		ch := make(chan *ollamaconnector.TaskResponse)
		respStr := []string{}
		go ollamaconnector.Ollamasettings.SendTaskStream(prompt, ch)
		for resp := range ch {
			respStr = append(respStr, resp.Response)
		}
		log.Printf("Kowalski: %s", strings.Join(respStr, ``))
		return nil
	},
	Args:       cobra.MinimumNArgs(1),
	ArgAliases: []string{"query", "ask", "q", "req", "r"},
	Hidden:     true,
}

func init() {
	chatCmd.AddCommand(reqCmd)
	chatCmd.PersistentFlags().String("location", "", "location of the actual files")
}

func GetCommand() *cobra.Command {
	return chatCmd
}
