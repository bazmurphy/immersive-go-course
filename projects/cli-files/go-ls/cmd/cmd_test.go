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
		flags            Flags
		args             []string
		testFiles        []string
		testFileContents []string
		testDirectories  []string
		expectedStdout   string
	}{
		{
			name:            "no files, no folders",
			flags:           Flags{},
			args:            []string{},
			testFiles:       []string{},
			testDirectories: []string{},
			expectedStdout:  "",
		},
		{
			name:            "3 files, no folders",
			flags:           Flags{},
			args:            []string{},
			testFiles:       []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testDirectories: []string{},
			expectedStdout:  "test-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
		},
		{
			name:            "no files, 3 folders",
			flags:           Flags{},
			args:            []string{},
			testFiles:       []string{},
			testDirectories: []string{"test-directory-1", "test-directory-2", "test-directory-3"},
			expectedStdout:  "test-directory-1\ntest-directory-2\ntest-directory-3\n",
		},
		{
			name:            "3 files, 3 folders",
			flags:           Flags{},
			args:            []string{},
			testFiles:       []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testDirectories: []string{"test-directory-1", "test-directory-2", "test-directory-3"},
			expectedStdout:  "test-directory-1\ntest-directory-2\ntest-directory-3\ntest-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
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

			// at the end of the test: change back to the original working directory
			defer func() {
				err = os.Chdir(originalWorkingDirectory)
				if err != nil {
					t.Fatalf("failed to change to the temporary directory: %v", err)
				}
			}()

			// store the original stdout
			originalStdout := os.Stdout

			// at the end of the test: restore the original stdout
			defer func() {
				os.Stdout = originalStdout
			}()

			// create a pipe to capture the output
			pipeRead, pipeWrite, _ := os.Pipe()

			// redirect stdout to the write end of the pipe
			os.Stdout = pipeWrite

			Execute(&testCase.flags, testCase.args)

			// close the write end of the pipe
			pipeWrite.Close()

			// read the captured output from the read end of the pipe
			pipeReadBytes, _ := io.ReadAll(pipeRead)

			actualStdout := string(pipeReadBytes)

			if actualStdout != testCase.expectedStdout {
				t.Errorf("output: actual %v | expected %v", actualStdout, testCase.expectedStdout)
			}
		})
	}
}
