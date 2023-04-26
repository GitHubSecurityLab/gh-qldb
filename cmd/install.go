package cmd

import (
  "io"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

  "github.com/spf13/cobra"
  "github.com/GitHubSecurityLab/gh-qldb/utils"
)

var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Install a local CodeQL database in the QLDB directory",
    Long:  `Install a local CodeQL database in the QLDB directory`,
    Run: func(cmd *cobra.Command, args []string) {
      install(nwoFlag, dbPathFlag, removeFlag)
    },
  }

func init() {
  rootCmd.AddCommand(installCmd)
  installCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO to associate the database to.")
  installCmd.Flags().StringVarP(&dbPathFlag, "database", "d", "", "The path to the database to install.")
  installCmd.Flags().BoolVarP(&removeFlag, "remove", "r", false, "Remove the database after installing it.")
  installCmd.MarkFlagRequired("nwo")
  installCmd.MarkFlagRequired("database")
}

func install(nwo string, dbPath string, remove bool) {
	fmt.Printf("Installing '%s' DB for '%s'\n", dbPath, nwo)

	// Check if the path exists
	fileinfo, err := os.Stat(dbPath)
	var zipPath string
	if os.IsNotExist(err) {
		log.Fatal(errors.New("DB path does not exist"))
	}
	if fileinfo.IsDir() {
		err := utils.ValidateDB(dbPath)
		if err != nil {
			fmt.Println("DB is not valid")
		}
		// Compress DB
		zipfilename := filepath.Join(os.TempDir(), "qldb.zip")
		fmt.Println("Zipping DB to", zipfilename)
		if err := utils.ZipDirectory(zipfilename, dbPath); err != nil {
			log.Fatal(err)
		}
		zipPath = zipfilename

	} else {
		// Check if the file is a zip
		if !strings.HasSuffix(dbPath, ".zip") {
			log.Fatal(errors.New("DB path is not a zip file"))
		}

		zipPath = dbPath
		// Unzip to temporary directory
		tmpdir, _ := ioutil.TempDir("", "qldb")
		_, err := utils.Unzip(dbPath, tmpdir)
		if err != nil {
			log.Fatal(err)
		}
		files, err := ioutil.ReadDir(tmpdir)
		if err != nil {
			log.Fatal(err)
		}
		if len(files) == 1 {
			tmpdir = filepath.Join(tmpdir, files[0].Name())
		}
		err = utils.ValidateDB(tmpdir)
		if err != nil {
			fmt.Println("DB is not valid")
		}
	}

	// read bytes from dbPath
	zipFile, err := os.Open(zipPath)
	if err != nil {
		log.Fatal(err)
	}
	defer zipFile.Close()
	zipBytes, err := ioutil.ReadAll(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	commitSha, primaryLanguage, err := utils.ExtractDBInfo(zipBytes)

	// Destination path
	dir := filepath.Join(utils.GetPath(nwo), primaryLanguage)
	filename := fmt.Sprintf("%s.zip", commitSha)
	path := filepath.Join(dir, filename)
	fmt.Println("Installing DB to", path)

	// Check if the DB is already installed
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		// Copy DB to the right place
		srcFile, err := os.Open(zipPath)
		if err != nil {
			log.Fatal(err)
		}
		defer srcFile.Close()
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			log.Fatal(err)
		}
		destFile, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer destFile.Close()
		fmt.Println("Copying DB to", path)
		_, err = io.Copy(srcFile, destFile) // check first var for number of bytes copied
		if err != nil {
			log.Fatal(err)
		}
		err = destFile.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}
	// Remove DB from the current location if -r flag is set
	if remove {
		fmt.Println("Removing DB from", dbPath)
		os.Remove(dbPath)
	}
}
