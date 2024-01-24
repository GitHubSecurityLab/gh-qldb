package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/cli/go-gh"
	graphql "github.com/shurcooL/githubv4"
)

const (
	ROOT = "codeql-dbs"
	VCS  = "github.com"
)

func GetBasePath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ROOT, VCS)
}

func GetPath(nwo string) string {
	return filepath.Join(GetBasePath(), nwo)
}

func ValidateDB(dbPath string) error {
	cmd := exec.Command("codeql", "resolve", "database", dbPath)
	cmd.Env = os.Environ()
	jsonBytes, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	var info interface{}
	err = json.Unmarshal(jsonBytes, &info)
	if err != nil {
		return err
	}
	return nil
}

func ExtractDBInfo(body []byte) (map[string]interface{}, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Extracting database information ... ")
	for _, zf := range zipReader.File {
		if strings.HasSuffix(zf.Name, "codeql-database.yml") {
			f, err := zf.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			yamlBytes, err := io.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			var dbData map[string]interface{}
			err = yaml.Unmarshal(yamlBytes, &dbData)
			if err != nil {
				log.Fatal(err)
			}
			return dbData, nil
		}
	}
	return nil, errors.New("codeql-database.yml not found")
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// ZipDirectory compresses a directory into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: directory add to the zip.
func ZipDirectory(zipFileName string, directoryToZip string) error {

	info, err := os.Stat(directoryToZip)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New(directoryToZip + " is not a directory")
	}

	// Create a new zip file
	newZipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	// Create a new zip archive
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Walk through the directory to zip and add each file to the archive
	err = filepath.Walk(directoryToZip, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}

		// Create a new file header for the file
		fileHeader, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}

		// Set the name of the file header to the relative path of the file
		// get parent directory name
		relPath, err := filepath.Rel(directoryToZip, filePath)
		if err != nil {
			return err
		}
		fileHeader.Name = filepath.Join("codeql-db", relPath)

		// Add the file header to the archive
		writer, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return err
		}

		// Open the file for reading
		fileToZip, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		// Copy the file to the archive
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Successfully created zip file %s\n", zipFileName)

	return nil
}

func GetCommitInfo(nwo string, commitSha string) (string, string, error) {

	graphqlClient, err := gh.GQLClient(nil)
	if err != nil {
		return "", "", err
	}
	var query struct {
		Repository struct {
			Object struct {
				Commit struct {
					AbbreviatedOid graphql.String
					CommittedDate  graphql.DateTime
				} `graphql:"... on Commit"`
			} `graphql:"object(oid: $sha)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": graphql.String(strings.Split(nwo, "/")[0]),
		"name":  graphql.String(strings.Split(nwo, "/")[1]),
		"sha":   graphql.GitObjectID(commitSha),
	}
	err = graphqlClient.Query("CommitInfo", &query, variables)
	if err != nil {
		return "", "", err
	}
	return string(query.Repository.Object.Commit.AbbreviatedOid), query.Repository.Object.Commit.CommittedDate.Format("2006-01-02T15:04:05"), nil
}

func GetCommitInfo2(nwo string, commitSha string) (string, string, error) {
	restClient, err := gh.RESTClient(nil)
	response := make(map[string]interface{})
	err = restClient.Get(fmt.Sprintf("repos/%s/commits/%s", nwo, commitSha), &response)
	if err != nil {
		log.Fatal(err)
	}
	commitMap := response["commit"].(map[string]interface{})
	committerMap := commitMap["committer"].(map[string]interface{})
	dateString := committerMap["date"].(string)
	return commitSha, dateString, nil
}
