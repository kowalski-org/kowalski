package cmd

import (
	"fmt"
	"os"

	chatcmd "github.com/openSUSE/kowalski/cmd/chat"
	databasecmd "github.com/openSUSE/kowalski/cmd/database"
	evaluatecmd "github.com/openSUSE/kowalski/cmd/evaluate"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/database"
	"github.com/openSUSE/kowalski/internal/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/charmbracelet/log"
)

var cfgFile string
var logLevel int

// rootCmd represents the base command when called without any subcommands
func RootCmd() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "kowalski",
		Short: "Helper for configuring your computer",
		Long: `Setup anything based on files with the help of
ollama and a knowledge database created from
distribution documentation.`,
		Run: func(cmd *cobra.Command, args []string) {
			if ok, _ := cmd.Flags().GetBool("version"); ok {
				os.Exit(0)
			} else {
				cmd.Usage()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				log.SetLevel(log.DebugLevel)
			}
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				conf_name := f.Name
				if !f.Changed && viper.IsSet(conf_name) {
					val := viper.Get(conf_name)
					cmd.Flags().Set(conf_name, fmt.Sprintf("%v", val))
				}
			})
		},
	}
	rootCmd.PersistentFlags().StringVar(&ollamaconnector.Ollamasettings.LLM, "modell", "gemma3:4b", "LLM modell to be used for answers")
	rootCmd.PersistentFlags().StringVar(&ollamaconnector.Ollamasettings.EmbeddingModel, "embedding", "nomic-embed-text", "embedding model for the knowledge database")
	rootCmd.PersistentFlags().StringVar(&ollamaconnector.Ollamasettings.OllamaURL, "url", "http://localhost:11434", "base URL for ollama requests")
	rootCmd.PersistentFlags().StringVar(&database.DBLocation, "database", "/usr/lib/kowalski", "path to knowledge database")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "turn on debugging messages")
	// viper.BindPFlags(rootCmd.PersistentFlags())
	// when this action is called directly.
	rootCmd.AddCommand(chatcmd.GetCommand())
	rootCmd.AddCommand(databasecmd.GetCommand())
	rootCmd.AddCommand(evaluatecmd.GetCommand())
	rootCmd.AddCommand(versCmd)
	rootCmd.Flags().BoolP("version", "v", false, "print version (git tag)")
	return rootCmd
}

func Execute() {
	err := RootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kowalski" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".config/kowalski")
	}

	viper.SetEnvPrefix("kw")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

var versCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Run: func(cmd *cobra.Command, args []string) {
		printVers()
	},
}

func printVers() {
	fmt.Fprintf(os.Stdout, "kowalksi version : %s", version.Version)
}
