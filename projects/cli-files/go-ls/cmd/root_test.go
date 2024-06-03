package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExec(t *testing.T) {
	t.Run("go-ls on a directory with a few files", func(t *testing.T) {
		// make a temporary directory (deliberately not using os.MkdirTemp)
		temporaryDirectoryName := "temp"
		temporaryDirectoryPath := filepath.Join(".", temporaryDirectoryName)
		err := os.Mkdir("temp", os.ModePerm)
		if err != nil {
			t.Errorf("error creating the temporary directory")
		}

		// create 3 new temporary txt files in that directory (deliberately not using os.CreateTemp)
		for i := 0; i < 3; i++ {
			temporaryFileName := fmt.Sprintf("test_file_%d.txt", i)
			temporaryFilePath := filepath.Join(temporaryDirectoryPath, temporaryFileName)
			temporaryFile, err := os.Create(temporaryFilePath)
			if err != nil {
				t.Errorf("error creating the temporary file")
			}
			// remember to close the file after creating it
			temporaryFile.Close()
		}

		// remove the temporary directory after the test
		defer os.RemoveAll(temporaryDirectoryPath)

		// to do:
		// capture standard output and compare actual with expected

		Execute()
	})

	// t.Run("go-ls on a directory with the -h flag", func(t *testing.T) {
	// 	// to do
	// 	Execute()
	// })

	// t.Run("go-ls on no directory", func(t *testing.T) {
	// 	// to do
	// 	Execute()
	// })
}
