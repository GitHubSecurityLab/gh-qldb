package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/GitHubSecurityLab/gh-qldb/utils"
	"github.com/spf13/cobra"
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
		fmt.Printf("Validating %s DB\n", dbPath)
		err := utils.ValidateDB(dbPath)
		if err != nil {
			fmt.Println("DB is not valid")
		}
		// Compress DB
		zipfilename := filepath.Join(os.TempDir(), "qldb.zip")
		fmt.Println("Compressing DB to", zipfilename)
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
		// Unzip to a temporary directory
		tmpdir, _ := os.MkdirTemp("", "qldb")

		_, err := utils.Unzip(dbPath, tmpdir)
		if err != nil {
			log.Fatal(err)
		}

		// Read all files in the tmpdir directory using os.ReadDir
		dirEntries, err := os.ReadDir(tmpdir)
		if err != nil {
			log.Fatal(err)
		}
		if len(dirEntries) == 1 {
			// if there is one directory in the tmpdir, use that as the tmpdir
			tmpdir = filepath.Join(tmpdir, dirEntries[0].Name())
		}
		fmt.Printf("Validating %s DB\n", tmpdir)
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
	zipBytes, err := io.ReadAll(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	commitSha, primaryLanguage, err := utils.ExtractDBInfo(zipBytes)
	shortCommitSha := commitSha[:8]
	fmt.Println("Commit SHA:", commitSha)
	fmt.Println("Short Commit SHA:", shortCommitSha)
	fmt.Println("Primary language:", primaryLanguage)

	// Destination path
	filename := fmt.Sprintf("%s-%s.zip", primaryLanguage, shortCommitSha)
	destPath := filepath.Join(utils.GetPath(nwo), filename)
	fmt.Println("Installing DB to", destPath)

	// Check if the DB is already installed
	if _, err := os.Stat(destPath); errors.Is(err, os.ErrNotExist) {

		// Create the directory if it doesn't exist
		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			log.Fatal(err)
			return
		}

		// Copy file from zipPath to destPath
		srcFile, err := os.Open(zipPath)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer destFile.Close()

		bytes, err := io.Copy(destFile, srcFile)
		fmt.Println(fmt.Sprintf("Copied %d bytes", bytes))

		if err != nil {
			log.Fatal(err)
		}
		err = destFile.Sync()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("DB already installed for same commit")
	}
	// Remove DB from the current location if -r flag is set
	if remove {
		fmt.Println("Removing DB from", dbPath)
		if err := os.RemoveAll(dbPath); err != nil {
			log.Fatal(err)
		}
	}
}
