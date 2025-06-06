package databasecmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"

	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/docbook"
	"github.com/openSUSE/kowalski/internal/pkg/information"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
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

type inputFormat string

const (
	textIn inputFormat = "text"
	yamlIn inputFormat = "yaml"
	jsonIn inputFormat = "json"
	xmlIn  inputFormat = "xml"
)

func (f *inputFormat) String() string {
	return string(*f)
}

func (f *inputFormat) Set(str string) error {
	switch str {
	case "text", "yaml", "json", "xml":
		*f = inputFormat(str)
		return nil
	default:
		return fmt.Errorf("Unkown input format: %s", str)
	}
}

func (f *inputFormat) Type() string {
	return "infoFormat"
}

var iFormat inputFormat
var databaseAdd = &cobra.Command{
	Use:     "add DATABASE FILE(s)",
	Aliases: []string{"create", "ad", "new"},
	Short:   "Add document(s) to the given database",
	Long: `Add a document extracted from a file
to the given database and create embeddings for it.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		db, err := database.New()
		if err != nil {
			return err
		}
		embedding, err := database.GetEmbedding([]string{args[0]})
		if err != nil {
			return err
		}
		embeddingSize, err := ollamaconnector.Ollamasettings.GetEmbeddingSize(embedding)
		if err != nil {
			return err
		}
		switch iFormat {
		case xmlIn:
			for i := range args[1:] {
				info, err := docbook.ParseDocBook(args[i+1], embeddingSize)
				if err != nil {
					log.Warnf("couldn't read file: %s", err)
					continue
				}
				if !info.Empty() {
					err = db.AddInformation(args[0], info)
					if err != nil {
						log.Warnf("file %s couldn't be added: %s", args[i+1], err)
						continue
					}
				} else {
					log.Warnf("file was empty: %s", args[i+1])
				}
			}
			return nil
		case yamlIn:
			for i := range args[1:] {
				info, err := information.ReadCurated(args[i+1])
				if err != nil {
					log.Warnf("couldn't read file: %s", err)
					continue
				}
				err = db.AddInformation(args[0], info)
				if err != nil {
					log.Warnf("file %s couldn't be added: %s", args[i+1], err)
					continue
				}
			}
			return nil
		default:
			return fmt.Errorf("unknown input type")
		}

	},
	Annotations: map[string]string{},
}

var databaseList = &cobra.Command{
	Use:     "list DATABASE [queries]",
	Aliases: []string{"ls"},
	Short:   "List (all) documents in the database",
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
					fmt.Fprintf(w, "%s %s %d %d\n", doc.Id, doc.Source, doc.NrFiles, doc.NrCommands)
				}
				w.Flush()
			}
		}
		return nil
	},
}

type outputFormat string

const (
	titleOut outputFormat = "title"
	fullOut  outputFormat = "full"
	yamlOut  outputFormat = "yaml"
	jsonOut  outputFormat = "json"
)

func (f *outputFormat) String() string {
	return string(*f)
}

func (f *outputFormat) Set(str string) error {
	switch str {
	case "title", "full", "yaml", "json":
		*f = outputFormat(str)
		return nil
	default:
		return fmt.Errorf("Unkown output format: %s", str)
	}
}

func (f *outputFormat) Type() string {
	return "infoFormat"
}

var oFormat outputFormat
var databaseGet = &cobra.Command{
	Use:     "get ID",
	Aliases: []string{"show", "cat"},
	Short:   "Get the information with ID out of database",
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
		case fullOut:
			fmt.Println(info.Render(templates.RenderInfoWithMeta))
		case yamlOut:
			str, _ := yaml.Marshal(info)
			fmt.Println(string(str))
		case jsonOut:
			str, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(str))
		default:
			fmt.Println(info.Render(templates.RenderTitleOnly))
		}
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

var databaseCheck = &cobra.Command{
	Use:     "check db for question",
	Aliases: []string{"chk"},
	Short:   "Check if database has a entry near the question",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New()
		defer db.Close()
		if err != nil {
			return err
		}
		collections := []string{}
		if len(args) > 1 {
			collections = args[1:]
		} else {
			collections, err = db.GetCollections()
			if err != nil {
				return err
			}
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
			switch oFormat {
			case fullOut:
				fmt.Println(info.Render())
			case yamlOut:
				str, _ := yaml.Marshal(info)
				fmt.Println(string(str))
			case jsonOut:
				str, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(str))
			default:
				fmt.Printf("%s %s\n", info.Hash, info.Title)
			}
		}
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

var dropDocuments = &cobra.Command{
	Use:     "drop [DocumentId]",
	Short:   "drop documents with given id from database",
	Aliases: []string{"rm", "remove", "delete", "del"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.New()
		if err != nil {
			log.Warnf("db error: %s", err)
			return
		}
		for _, docId := range args {
			err = db.DropInformation(docId)
			if err != nil {
				log.Warn(err)
				return
			}
		}
	},
}

func init() {
	databaseGet.Flags().Var(&oFormat, "format", "format of the dump {full,title,json,yaml}")
	databaseCheck.Flags().Var(&oFormat, "format", "format of the dump {full,title,json,yaml}")
	databaseAdd.Flags().Var(&iFormat, "format", "format of the input {text,json,xml,yaml}")
	// need to set as Var hasn't a default input
	databaseAdd.Flags().Set("format", "xml")
	databaseCmd.AddCommand(databaseAdd)
	databaseCmd.AddCommand(databaseList)
	databaseCmd.AddCommand(databaseCheck)
	databaseCheck.Flags().Int64P("number", "n", 5, "number of documents to retreive")
	databaseCmd.AddCommand(databaseGet)
	databaseCmd.AddCommand(dropDocuments)
}
func GetCommand() *cobra.Command {
	return databaseCmd
}
