package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel log.Level

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dns-checker",
	Short: "Bulk dns redirect verifier.",
	Long:  `Verifies a batch of url's and their respective destinations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetLevel(logLevel)
		log.Debugf("Log Level: %v", logLevel)
	},
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dns-checker.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().Uint32VarP((*uint32)(&logLevel), "log-level", "l", 2, "Logging level: 0-6 the higher the more verbose.")
}
