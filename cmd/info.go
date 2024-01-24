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

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Returns information about a database stored in the QLDB structure",
	Long:  `Returns information about a database stored in the QLDB structure`,
	Run: func(cmd *cobra.Command, args []string) {
		info(nwoFlag, languageFlag)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to get the database for.")
	infoCmd.Flags().StringVarP(&languageFlag, "language", "l", "", "The primary language you want the database for.")
	infoCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Use json as the output format.")
	infoCmd.MarkFlagRequired("nwo")
	infoCmd.MarkFlagRequired("language")
}

func info(nwo string, language string) {
	dir := utils.GetPath(nwo)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var pathList []map[string]string
	for _, e := range entries {
		entryName := e.Name()
		var name string
		if filepath.Ext(entryName) == ".zip" {
			// remove the .zip extension if it exists
			name = entryName[:len(entryName)-len(filepath.Ext(entryName))]
		} else if e.IsDir() {
			name = e.Name()
		} else {
			continue
		}

		// split the name by the "-". first element is the language, second is the short commit sha
		nameSplit := strings.Split(name, "-")
		if len(nameSplit) != 2 {
			log.Fatal(fmt.Errorf("invalid database name: %s", name))
		}
		if nameSplit[0] != language {
			continue
		}
		shortSha := nameSplit[1]

		commitSha, committedDate, err := utils.GetCommitInfo2(nwo, shortSha)
		if err != nil {
			log.Fatal(err)
		}
		pathList = append(pathList, map[string]string{
			"commitSha":     commitSha,
			"committedDate": committedDate,
			"path":          filepath.Join(dir, entryName),
		})
	}
	if jsonFlag {
		jsonStr, err := json.MarshalIndent(pathList, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", jsonStr)
	} else {
		for _, path := range pathList {
			fmt.Println(path["path"])
		}
	}
}
