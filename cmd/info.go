package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	dir := filepath.Join(utils.GetPath(nwo), language)
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var pathList []map[string]string
	for _, f := range files {
		filename := f.Name()
		sha := filename[:len(filename)-len(filepath.Ext(filename))]
		commitSha, committedDate, err := utils.GetCommitInfo(nwo, sha)
		if err != nil {
			log.Fatal(err)
		}
		pathList = append(pathList, map[string]string{
			"commitSha":     commitSha,
			"committedDate": committedDate,
			"path":          filepath.Join(dir, filename),
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
