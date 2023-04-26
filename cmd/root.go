package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var ( 
  nwoFlag string
  languageFlag string
  removeFlag bool
  dbPathFlag string
  jsonFlag bool
)
var rootCmd = &cobra.Command{
  Use:   "gh-qldb",
  Short: "A CodeQL database manager",
  Long: `A CodeQL database manager. Download, deploy and create CodeQL databases with ease.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
