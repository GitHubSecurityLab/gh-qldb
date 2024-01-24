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
		info()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to get the database for.")
	infoCmd.Flags().StringVarP(&languageFlag, "language", "l", "", "The primary language you want the database for.")
	infoCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Use json as the output format.")
	infoCmd.Flags().StringVarP(&dbPathFlag, "db-path", "p", "", "Path to the database to get the info from.")
	infoCmd.MarkFlagsOneRequired("db-path", "nwo")
	infoCmd.MarkFlagsMutuallyExclusive("db-path", "nwo")
}

func info() {
	var results []map[string]string

	if nwoFlag != "" {
		results = infoFromNwo(nwoFlag)
	} else if dbPathFlag != "" {
		results = append(results, infoFromPath(dbPathFlag))
	}
	if jsonFlag {
		jsonStr, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", jsonStr)
	} else {
		for _, result := range results {
			fmt.Println(result["path"])
		}
	}
}

func infoFromPath(path string) map[string]string {
	// get the file name part of path
	parts := strings.Split(path, string(os.PathSeparator))
	name := parts[len(parts)-1]
	base := filepath.Dir(path)

	var dbname string
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		dbname = path
	} else if filepath.Ext(name) == ".zip" {
		dbname = name[:len(name)-len(filepath.Ext(name))]
	} else {
		log.Fatal(fmt.Errorf("invalid database path: %s", path))
		return nil
	}

	// split the name by the "-". first element is the language, second is the short commit sha
	nameSplit := strings.Split(dbname, "-")
	if len(nameSplit) != 2 {
		log.Fatal(fmt.Errorf("invalid database name: %s", name))
		return nil
	}

	lang := nameSplit[0]
	shortSha := nameSplit[1]
	baseParts := strings.Split(base, string(os.PathSeparator))
	nwo := filepath.Join(baseParts[len(baseParts)-2], baseParts[len(baseParts)-1])
	commitSha, committedDate, err := utils.GetCommitInfo2(nwo, shortSha)
	if err != nil {
		log.Fatal(err)
	}

	return map[string]string{
		"commitSha":     commitSha,
		"committedDate": committedDate,
		"language":      lang,
		"path":          path,
	}
}

func infoFromNwo(nwo string) []map[string]string {
	dir := utils.GetPath(nwo)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var pathList []string

	for _, e := range entries {
		entryName := e.Name()
		if filepath.Ext(entryName) == ".zip" || e.IsDir() {
			pathList = append(pathList, filepath.Join(dir, entryName))
		}
	}

	var results []map[string]string
	for _, path := range pathList {
		results = append(results, infoFromPath(path))
	}
	return results
}
