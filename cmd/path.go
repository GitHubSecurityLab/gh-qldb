package cmd

import (
  "github.com/GitHubSecurityLab/gh-qldb/utils"
  "github.com/spf13/cobra"
  "path/filepath"
  "log"
  "fmt"
  "io/ioutil"
  "encoding/json"
)
var pathCmd = &cobra.Command{
    Use:   "path",
    Short: "Returns the path to a database stored in the QLDB structure",
    Long:  `Returns the path to a database stored in the QLDB structure`,
    Run: func(cmd *cobra.Command, args []string) {
      path(nwoFlag, languageFlag)
    },
  }

func init() {
  rootCmd.AddCommand(pathCmd)
  pathCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to get the database for.")
  pathCmd.Flags().StringVarP(&languageFlag, "language", "l", "", "The primary language you want the database for.")
  pathCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Use json as the output format.")
  pathCmd.MarkFlagRequired("nwo")
  pathCmd.MarkFlagRequired("language")
}

func path(nwo string, language string) {
  dir := filepath.Join(utils.GetPath(nwo), language)
  files, err := ioutil.ReadDir(dir)
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
      "commitSha": commitSha,
      "committedDate": committedDate,
      "path": filepath.Join(dir, filename),
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
        fmt.Printf("%s (%s)", path["path"], path["committedDate"])
      }
    }
}

