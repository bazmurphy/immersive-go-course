// projects/cli-files/go-ls/cmd/root_test.go

package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestExecute(t *testing.T) {
	testCases := []struct {
		name             string
		testFiles        []string
		testFileContents []string
		testDirectories  []string
		expectedOutput   string
	}{
		{
			name:            "no files, no folders",
			testFiles:       []string{},
			testDirectories: []string{},
			expectedOutput:  "",
		},
		{
			name:            "3 files, no folders",
			testFiles:       []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testDirectories: []string{},
			expectedOutput:  "test-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
		},
		{
			name:            "no files, 3 folders",
			testFiles:       []string{},
			testDirectories: []string{"test-directory-1", "test-directory-2", "test-directory-3"},
			expectedOutput:  "test-directory-1\ntest-directory-2\ntest-directory-3\n",
		},
		{
			name:            "3 files, 3 folders",
			testFiles:       []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testDirectories: []string{"test-directory-1", "test-directory-2", "test-directory-3"},
			expectedOutput:  "test-directory-1\ntest-directory-2\ntest-directory-3\ntest-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// get the original working directory
			originalWorkingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to store original working directory: %v", err)
			}

			// create a temporary directory to create the test directories and test files in
			temporaryDirectory, err := os.MkdirTemp("", "go-ls-temporary-directory")
			if err != nil {
				t.Fatalf("failed to create a temporary directory: %v", err)
			}
			defer os.RemoveAll(temporaryDirectory)

			// create the test directories
			for _, testDirectory := range testCase.testDirectories {
				testDirectoryPath := filepath.Join(temporaryDirectory, testDirectory)
				err := os.MkdirAll(testDirectoryPath, os.ModePerm)
				if err != nil {
					t.Fatalf("failed to create test directory: %v", err)
				}
			}

			// create the test files
			for _, testFile := range testCase.testFiles {
				testFilePath := filepath.Join(temporaryDirectory, testFile)
				createdTestFile, err := os.Create(testFilePath)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				createdTestFile.Close()
			}

			// change to the temporary directory to run Execute()
			err = os.Chdir(temporaryDirectory)
			if err != nil {
				t.Fatalf("failed to change to the temporary directory: %v", err)
			}

			defer func() {
				// at the end of the test: change back to the original working directory
				err = os.Chdir(originalWorkingDirectory)
				if err != nil {
					t.Fatalf("failed to change to the temporary directory: %v", err)
				}
			}()

			// ---------- I found this difficult to fully grasp ----------

			// store the original stdout
			originalStdout := os.Stdout

			defer func() {
				// at the end of the test: restore the original stdout
				os.Stdout = originalStdout
			}()

			// create a pipe to capture the output
			pipeRead, pipeWrite, _ := os.Pipe()

			// redirect stdout to the write end of the pipe
			os.Stdout = pipeWrite

			Execute()

			// close the write end of the pipe
			pipeWrite.Close()

			// read the captured output from the read end of the pipe
			pipeReadBytes, _ := io.ReadAll(pipeRead)
			// and this is bad because i am reading the whole thing at once and not streaming it...
			// which totally defeats the point of reading it line by line earlier...

			actualOutput := string(pipeReadBytes)

			// -----------------------------------------------------------

			if actualOutput != testCase.expectedOutput {
				t.Errorf("output: actual %v | expected %v", actualOutput, testCase.expectedOutput)
			}
		})
	}
}
