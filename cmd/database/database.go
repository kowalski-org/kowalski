package databasecmd

import (
	"maps"

	"github.com/mslacken/kowalski/internal/pkg/database"
	"github.com/mslacken/kowalski/internal/pkg/docbook"
	"github.com/spf13/cobra"
)

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage the different databases",
	Long: `List, create databse from sources or
permanenetly remove databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var databaseAdd = &cobra.Command{
	Use:        "add DATABASE FILE(s)",
	ArgAliases: []string{"create", "ad", "new"},
	Short:      "Add document(s) to the given database",
	Long: `Add a document extracted from a file
to the given database and create embeddings for it.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		entities, err := cmd.PersistentFlags().GetStringArray("entity")
		if err != nil {
			return err
		}
		entitiesMap := make(map[string]string)
		for _, ent := range entities {
			entMap, err := docbook.ReadEntity(ent)
			if err != nil {
				return err
			}
			maps.Copy(entitiesMap, entMap)
		}
		db := database.New()
		for i := range args[1:] {
			bk := docbook.Docbook{
				Entities: entitiesMap,
			}
			info, err := bk.ParseDocBook(args[i+1])
			err = db.AddInformation(args[0], info)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	databaseCmd.AddCommand(databaseAdd)
	databaseAdd.PersistentFlags().StringArray("entity", []string{}, "filename of an xml entity defintions")
}
func GetCommand() *cobra.Command {
	return databaseCmd
}
