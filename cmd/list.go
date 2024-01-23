package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/GitHubSecurityLab/gh-qldb/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of CodeQL databases stored in the QLDB structure",
	Long:  `Returns a list of CodeQL databases stored in the QLDB structure`,
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to get the databases for.")
	listCmd.Flags().StringVarP(&languageFlag, "language", "l", "", "The primary language you want the databases for.")
	listCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Use json as the output format.")
}

func list() {
	var results []string
	basePath := utils.GetBasePath()
	dirEntries, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			user := dirEntry.Name()
			userPath := filepath.Join(basePath, user)
			repoEntries, err := os.ReadDir(userPath)
			if err != nil {
				log.Fatal(err)
			}
			for _, repoEntry := range repoEntries {
				if repoEntry.IsDir() {
					repo := repoEntry.Name()
					nwoPath := filepath.Join(userPath, repo)
					dbEntries, err := os.ReadDir(nwoPath)
					if err != nil {
						log.Fatal(err)
					}
					for _, dbEntry := range dbEntries {
						if dbEntry.IsDir() || filepath.Ext(dbEntry.Name()) == ".zip" {
							dbPath := filepath.Join(nwoPath, dbEntry.Name())
							results = append(results, dbPath)
						}
					}
				}
			}
		}
	}

	// if languageFlag is set, filter the results
	// so that the entries in results only contains elements which filename starts with languageFlag-
	if languageFlag != "" {
		var filteredResults []string
		for _, result := range results {
			if strings.HasPrefix(filepath.Base(result), languageFlag+"-") {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// if nwoFlag is set, filter the results so that the results only contains elements which contains nwoFlag
	if nwoFlag != "" {
		var filteredResults []string
		for _, result := range results {
			if strings.Contains(strings.ToLower(result), strings.ToLower(nwoFlag)) {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// if jsonFlag is set, print the results as json
	if jsonFlag {
		jsonBytes, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		for _, result := range results {
			fmt.Println(result)
		}
	}

}
