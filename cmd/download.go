package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/GitHubSecurityLab/gh-qldb/utils"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads a CodeQL database from GitHub Code Scanning",
	Long:  `Downloads a CodeQL database from GitHub Code Scanning`,
	Run: func(cmd *cobra.Command, args []string) {
		download()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to download the database for.")
	downloadCmd.Flags().StringVarP(&languageFlag, "language", "l", "", "The primary language you want the database for.")
	downloadCmd.MarkFlagRequired("nwo")
	downloadCmd.MarkFlagRequired("language")

}

func download() {
	// fetch the DB info from GitHub API
	fmt.Printf("Fetching DB info for '%s'\n", nwoFlag)
	restClient, err := gh.RESTClient(nil)
	response := make([]interface{}, 0)
	err = restClient.Get(fmt.Sprintf("repos/%s/code-scanning/codeql/databases", nwoFlag), &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Found DBs for the following languages: ")
	for i, v := range response {
		dbMap := v.(map[string]interface{})
		language := dbMap["language"].(string)
		fmt.Print(language)
		if i < len(response)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Print("\n")

	// download the DBs
	for _, v := range response {
		dbMap := v.(map[string]interface{})
		language := dbMap["language"].(string)
		if languageFlag != "all" && language != languageFlag {
			continue
		}

		// download DB
		fmt.Printf("Downloading '%s' DB for '%s'\n", language, nwoFlag)
		opts := api.ClientOptions{
			Headers: map[string]string{"Accept": "application/zip"},
		}
		httpClient, err := gh.HTTPClient(&opts)
		if err != nil {
			log.Fatal(err)
		}
		url := fmt.Sprintf("https://api.github.com/repos/%s/code-scanning/codeql/databases/%s", nwoFlag, language)
		resp, err := httpClient.Get(url)
		if err != nil {
			fmt.Printf("Failure downloading the DB from `%s`: %v", url, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// get the commit this DB was created from
		metadata, err := utils.ExtractDBInfo(body)
		if err != nil {
			log.Fatal(err)
		}
		metadata["provenance"] = nwoFlag
		commitSha := metadata["creationMetadata"].(map[string]interface{})["sha"].(string)
		shortCommitSha := commitSha[:8]
		primaryLanguage := metadata["primaryLanguage"].(string)
		fmt.Println()
		fmt.Println("Commit SHA:", commitSha)
		fmt.Println("Short Commit SHA:", shortCommitSha)
		fmt.Println("Primary language:", primaryLanguage)

		zipFilename := fmt.Sprintf("%s-%s.zip", primaryLanguage, shortCommitSha)
		jsonFilename := fmt.Sprintf("%s-%s.json", primaryLanguage, shortCommitSha)
		dir := utils.GetPath(nwoFlag)
		zipPath := filepath.Join(dir, zipFilename)
		jsonPath := filepath.Join(dir, jsonFilename)

		// create directory if not exists
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}

		// create DB file if doesnot exists
		if _, err := os.Stat(zipPath); errors.Is(err, os.ErrNotExist) {
			// write the DB to disk
			err = os.WriteFile(zipPath, body, 0755)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Writing DB to %s\n", zipPath)

		} else {
			fmt.Printf("Aborting, DB %s already exists\n", zipPath)
		}

		// create Metadata file if doesnot exists
		if _, err := os.Stat(jsonPath); errors.Is(err, os.ErrNotExist) {
			// Convert the map to JSON
			jsonData, err := json.Marshal(metadata)
			if err != nil {
				log.Fatal(err)
			}
			// Write the JSON data to a file
			err = os.WriteFile(jsonPath, jsonData, 0644)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("Aborting, DB metadata %s already exists\n", jsonPath)
		}
	}
	fmt.Println("Done")
}
