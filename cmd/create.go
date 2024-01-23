package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Extracts a CodeQL database from a source path",
	Long: `Extracts a CodeQL database from a source path. Pass the CodeQL arguments after a '--' separator.

eg: gh-qldb create --nwo foo/bar -- -s /path/to/src -l javascript`,
	Run: func(cmd *cobra.Command, args []string) {
		// --nwo foo/bar -- -s /path/to/src -l javascript
		create(nwoFlag, args)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&nwoFlag, "nwo", "n", "", "The NWO of the repository to create the database for. If omitted, it will be inferred from git remotes.")
}

func create(nwo string, codeqlArgs []string) {
	fmt.Printf("Creating DB for '%s'. CodeQL args: '%v'", nwo, codeqlArgs)
	destPath := filepath.Join(os.TempDir(), "codeql-db")
	if err := os.MkdirAll(destPath, 0755); err != nil {
		log.Fatal(err)
	}
	args := []string{"database", "create"}
	args = append(args, codeqlArgs...)
	args = append(args, "--")
	args = append(args, destPath)
	cmd := exec.Command("codeql", args...)
	cmd.Env = os.Environ()
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln(err)
	}

	install(nwo, destPath, true)
}
