package databasecmd

import (
	"fmt"

	"github.com/mslacken/kowalski/internal/pkg/database"
	"github.com/spf13/cobra"
)

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage the different databases",
	Long: `List, create databse from sources or
permanenetly remove databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("database called")
	},
}

var databaseAdd = &cobra.Command{
	Use:        "add DATABASE FILE(s)",
	ArgAliases: []string{"create", "ad", "new"},
	Short:      "Add document(s) to the given database",
	Long: `Add a document to the given database and create embeddings for it.
If a directory is given all documents in the directory are added.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := database.New()
		db.AddInformation("test",
			database.Information{
				OS:      "linux",
				Title:   "Test doc",
				Content: "This is a a simple test",
			},
		)
	},
}

func init() {
	databaseCmd.AddCommand(databaseAdd)
}
func GetCommand() *cobra.Command {
	return databaseCmd
}
