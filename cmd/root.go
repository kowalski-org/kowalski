package cmd

import (
	"fmt"
	"os"

	chatcmd "github.com/openSUSE/kowalski/cmd/chat"
	databasecmd "github.com/openSUSE/kowalski/cmd/database"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/charmbracelet/log"
)

var cfgFile string
var ollamaSettings ollamaconnector.Settings
var logLevel int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kowalski",
	Short: "Helper for configuring your computer",
	Long: `Setup anything based on files with the help of
ollama and a knowledge database created from
distribution documentation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kowalski.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.AddCommand(chatcmd.GetCommand())
	rootCmd.AddCommand(databasecmd.GetCommand())
	// modelname, _ := cmd.PersistentFlags().GetString("model")
	viper.SetDefault("llm", "gemma3:4b")
	viper.SetDefault("embedding", "nomic-embed-text")
	viper.SetDefault("URL", "http://localhost:11434/api/")
	rootCmd.PersistentFlags().String("modell", viper.GetString("llm"), "LLM modell to be used for answers")
	rootCmd.PersistentFlags().String("embedding", viper.GetString("embedding"), "embedding model for the knowledge database")
	rootCmd.PersistentFlags().String("URL", viper.GetString("URL"), "base URL for ollama requests")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "turn on debugging messages")
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
		viper.SetConfigName(".kowalski")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if debug, _ := rootCmd.Flags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
	}
}
