package databasecmd

import (
	"fmt"
	"github.com/charmbracelet/log"
	"maps"

	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/docbook"
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
	Args: cobra.MinimumNArgs(1),
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
		if dump, _ := cmd.PersistentFlags().GetBool("dumpentity"); dump {
			for key, val := range entitiesMap {
				fmt.Printf("%s: %s\n", key, val)
			}
			return nil
		}
		cmd.Args = cobra.MinimumNArgs(2)
		db, err := database.New()
		if err != nil {
			return err
		}
		for i := range args[1:] {
			bk := docbook.Docbook{
				Entities: entitiesMap,
			}
			info, err := bk.ParseDocBook(args[i+1])
			if err != nil {
				log.Printf("error on file: %s %s\n", args[i+1], err)
			}
			if !info.Empty() {
				err = db.AddInformation(args[0], info)
				if err != nil {
					return err
				}
			} else {
				log.Printf("file was empty: %s", args[i+1])
			}
		}
		return nil
	},
	Annotations: map[string]string{},
}

var databaseList = &cobra.Command{
	Use:        "list DATABASE [queries]",
	ArgAliases: []string{"ls"},
	Short:      "List (all) documents in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			if colls, err := db.ListCollections(); err != nil {
				return err
			} else {
				log.Info("Collections:\n")
				for _, col := range colls {
					log.Infof("%s\n", col)
				}
			}
		} else {
			if docs, err := db.List(args[0]); err != nil {
				return err
			} else {
				log.Infof("Documents:\n")
				for _, doc := range docs {
					log.Infof("%s\n", doc)
				}
			}
		}
		return nil
	},
}

var databaseCheck = &cobra.Command{
	Use:        "check db for question",
	ArgAliases: []string{"chk"},
	Short:      "Check if database has a entry near the question",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		defer db.Close()
		if err != nil {
			return err
		}
		collections := []string{}
		if len(args) > 1 {
			collections = args[1:]
		}
		infos, err := db.GetInfos(args[0], collections)
		if err != nil {
			return err
		}
		for _, info := range infos {
			fmt.Println(info.Render())
		}
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	databaseCmd.AddCommand(databaseAdd)
	databaseCmd.AddCommand(databaseList)
	databaseCmd.AddCommand(databaseCheck)
	databaseAdd.PersistentFlags().StringArray("entity", []string{}, "filename of an xml entity defintions")
	databaseAdd.PersistentFlags().Bool("dumpentity", false, "just dump the used entity map")
}
func GetCommand() *cobra.Command {
	return databaseCmd
}
