package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
  "github.com/spf13/cobra"
  "github.com/GitHubSecurityLab/gh-qldb/utils"
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

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// get the commit this DB was created from
		commitSha, primaryLanguage, err := utils.ExtractDBInfo(body)
		if err != nil {
			log.Fatal(err)
		}

		filename := fmt.Sprintf("%s.zip", commitSha)
		dir := filepath.Join(utils.GetPath(nwoFlag), primaryLanguage)
		path := filepath.Join(dir, filename)

		// create directory if not exists
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}

		// create file if not exists
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			// write the DB to disk
			err = ioutil.WriteFile(path, body, 0755)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Writing DB to %s\n", path)

		} else {
			fmt.Printf("Aborting, DB %s already exists\n", path)
		}

	}
	fmt.Println("Done")
}

