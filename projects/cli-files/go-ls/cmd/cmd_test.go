package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestExecute(t *testing.T) {
	testCases := []struct {
		name                    string
		flags                   Flags
		args                    []string
		testFiles               []string
		testSubdirectories      []string
		testSubdirectoriesFiles []string
		expectedStdout          string
		expectedStderr          string
	}{
		{
			name:           "an empty directory",
			expectedStdout: "",
		},
		{
			name:           "with the help flag",
			flags:          Flags{Help: true},
			expectedStdout: "go-ls: help message",
		},
		{
			name:           "with the all flag",
			flags:          Flags{All: true},
			testFiles:      []string{"test-file-1.txt", ".test-hidden-file"},
			expectedStdout: ".test-hidden-file\ntest-file-1.txt\n",
		},
		{
			name:           "a directory with 3 files",
			testFiles:      []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			expectedStdout: "test-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
		},
		{
			name:               "a directory with 3 subdirectories",
			testSubdirectories: []string{"test-subdirectory-1", "test-subdirectory-2", "test-subdirectory-3"},
			expectedStdout:     "test-subdirectory-1\ntest-subdirectory-2\ntest-subdirectory-3\n",
		},
		{
			name:               "a directory with 3 subdirectories and 3 files",
			testFiles:          []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testSubdirectories: []string{"test-subdirectory-1", "test-subdirectory-2", "test-subdirectory-3"},
			expectedStdout:     "test-subdirectory-1\ntest-subdirectory-2\ntest-subdirectory-3\ntest-file-1.txt\ntest-file-2.txt\ntest-file-3.txt\n",
		},
		{
			name:               "a non-existant directory",
			args:               []string{"non-existant-directory"},
			testFiles:          []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testSubdirectories: []string{"test-directory-1", "test-directory-2", "test-directory-3"},
			expectedStderr:     "go-ls: could not read the file/directory information: stat non-existant-directory: no such file or directory",
		},
		{
			name:               "a subdirectory and a non-existant directory",
			args:               []string{".", "non-existant-directory"},
			testFiles:          []string{"test-file-1.txt"},
			testSubdirectories: []string{"test-subdirectory-1"},
			expectedStdout:     ".:\ntest-subdirectory-1\ntest-file-1.txt\n\n",
			expectedStderr:     "go-ls: could not read the file/directory information: stat non-existant-directory: no such file or directory",
		},
		{
			name:                    "two subdirectories with files",
			args:                    []string{"test-directory-1", "test-directory-2"},
			testSubdirectories:      []string{"test-directory-1", "test-directory-2"},
			testSubdirectoriesFiles: []string{"test-subdirectory-1-file-1.txt", "test-subdirectory-2-file-1.txt"},
			expectedStdout:          "test-directory-1:\ntest-subdirectory-1-file-1.txt\n\ntest-directory-2:\ntest-subdirectory-2-file-1\n",
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

			if len(testCase.testSubdirectories) > 0 {
				// create the test subdirectories
				for index, testDirectory := range testCase.testSubdirectories {
					testDirectoryPath := filepath.Join(temporaryDirectory, testDirectory)
					err := os.MkdirAll(testDirectoryPath, os.ModePerm)
					if err != nil {
						t.Fatalf("failed to create test subdirectory: %v", err)
					}
					// create the subdirectory test files
					if len(testCase.testSubdirectoriesFiles) > 0 {
						testSubdirectoryFilePath := filepath.Join(testDirectoryPath, testCase.testSubdirectoriesFiles[index])
						createdSubdirectoryTestFile, err := os.Create(testSubdirectoryFilePath)
						if err != nil {
							t.Fatalf("failed to create test subdirectory file: %v", err)
						}
						createdSubdirectoryTestFile.Close()
					}
				}
			}

			if len(testCase.testFiles) > 0 {
				// create the test files
				for _, testFile := range testCase.testFiles {
					testFilePath := filepath.Join(temporaryDirectory, testFile)
					createdTestFile, err := os.Create(testFilePath)
					if err != nil {
						t.Fatalf("failed to create test file: %v", err)
					}
					createdTestFile.Close()
				}
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
					t.Fatalf("failed to change to the original directory: %v", err)
				}
			}()

			// (!!!) HIJACKING STDOUT AND STDERR IS BAD but i cannot figure out how to do it with bytes.Buffer like in main_test.go

			// store the original stdout
			originalStdout := os.Stdout
			originalStderr := os.Stderr

			// create a pipe to capture stdout/stderr
			stdoutPipeRead, stdoutPipeWrite, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create stdout pipe: %v", err)
			}

			stderrPipeRead, stderrPipeWrite, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create stderr pipe: %v", err)
			}

			// redirect stdout to the write end of the pipe
			os.Stdout = stdoutPipeWrite
			os.Stderr = stderrPipeWrite

			// at the end of the test: restore the original stdout
			defer func() {
				os.Stdout = originalStdout
				os.Stderr = originalStderr
			}()

			Execute(&testCase.flags, testCase.args)

			// close the write end of the pipe
			stdoutPipeWrite.Close()
			if err != nil {
				t.Fatalf("failed to close stdout pipe: %v", err)
			}

			stderrPipeWrite.Close()
			if err != nil {
				t.Fatalf("failed to close stderr pipe: %v", err)
			}

			// read the captured output from the read end of the pipe
			stdoutPipeReadBytes, err := io.ReadAll(stdoutPipeRead)
			if err != nil {
				t.Fatalf("failed to read from stdout pipe: %v", err)
			}

			stdoutPipeReadErr, err := io.ReadAll(stderrPipeRead)
			if err != nil {
				t.Fatalf("failed to read from stderr pipe: %v", err)
			}

			actualStdout := string(stdoutPipeReadBytes)
			actualStderr := string(stdoutPipeReadErr)

			if actualStdout != testCase.expectedStdout {
				t.Errorf("stdout: actual %v | expected %v", actualStdout, testCase.expectedStdout)
			}

			if actualStderr != testCase.expectedStderr {
				t.Errorf("stderr: actual %v | expected %v", actualStderr, testCase.expectedStderr)
			}
		})
	}
}
