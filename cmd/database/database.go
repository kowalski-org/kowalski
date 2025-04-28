package databasecmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/log"

	jsonpkg "encoding/json"

	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/docbook"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
	"github.com/spf13/cobra"
	yamlpkg "gopkg.in/yaml.v3"
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
		cmd.Args = cobra.MinimumNArgs(2)
		db, err := database.New()
		if err != nil {
			return err
		}
		for i := range args[1:] {
			info, err := docbook.ParseDocBook(args[i+1])
			if err != nil {
				return err
			}
			if !info.Empty() {
				err = db.AddInformation(args[0], info)
				if err != nil {
					return err
				}
			} else {
				log.Warnf("file was empty: %s", args[i+1])
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
				fmt.Println("Collections:")
				for _, col := range colls {
					fmt.Printf("%s\n", col)
				}
			}
		} else {
			if docs, err := db.List(args[0]); err != nil {
				return err
			} else {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
				fmt.Fprintln(w, "Id\tSource\tFiles\tCommands")
				for _, doc := range docs {

					fmt.Printf("%s %s %d %d\n", doc.Id, doc.Source, doc.NrFiles, doc.NrCommands)
				}
				w.Flush()
			}
		}
		return nil
	},
}

type infoFormat string

const (
	title infoFormat = "title"
	full  infoFormat = "full"
	yaml  infoFormat = "yaml"
	json  infoFormat = "json"
)

func (f *infoFormat) String() string {
	return string(*f)
}

func (f *infoFormat) Set(str string) error {
	switch str {
	case "title", "full", "yaml", "json":
		*f = infoFormat(str)
		return nil
	default:
		return fmt.Errorf("Unkown output format: %s", str)
	}
}

func (f *infoFormat) Type() string {
	return "infoFormat"
}

var oFormat infoFormat

var databaseGet = &cobra.Command{
	Use:        "get ID",
	ArgAliases: []string{"show", "cat"},
	Short:      "Get the information with ID out of database",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		if err != nil {
			return err
		}
		info, err := db.Get(args[0])
		if err != nil {
			return err
		}
		switch oFormat {
		case full:
			fmt.Println(info.Render(templates.RenderInfoWithMeta))
		case yaml:
			str, _ := yamlpkg.Marshal(info)
			fmt.Println(string(str))
		case json:
			str, _ := jsonpkg.MarshalIndent(info, "", "  ")
			fmt.Println(string(str))
		default:
			fmt.Println(info.Render(templates.RenderTitleOnly))
		}
		return nil
	},
	Args: cobra.MinimumNArgs(1),
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
		nrDocs, err := cmd.Flags().GetInt64("number")
		if err != nil {
			return err
		}
		infos, err := db.GetInfos(args[0], collections, nrDocs)
		if err != nil {
			return err
		}
		fmt.Println("Infos:")
		for _, info := range infos {
			str, _ := info.Render()
			fmt.Println(str)
		}
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	databaseGet.Flags().Var(&oFormat, "format", "format of the dump")
	databaseCmd.AddCommand(databaseAdd)
	databaseCmd.AddCommand(databaseList)
	databaseCmd.AddCommand(databaseCheck)
	databaseCheck.Flags().Int64P("number", "n", 5, "number of documents to retreive")
	databaseCmd.AddCommand(databaseGet)
}
func GetCommand() *cobra.Command {
	return databaseCmd
}
