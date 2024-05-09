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
		expectedOutput   string
		expectedError    error
	}{
		{
			name:             "no flags, args: 1 file, with a single line, ending WITH a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\n"},
			expectedOutput:   "hello\n",
		},
		{
			name:             "no flags, args: 1 file, with a single line, ending WITHOUT a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello"},
			expectedOutput:   "hello",
		},
		{
			name:             "-n flag, args: 1 file, with a single line, ending WITH a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\n"},
			expectedOutput:   "1\thello\n",
		},
		{
			name:             "-n flag, args: 1 file, with a single line, ending WITHOUT a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello"},
			expectedOutput:   "1\thello",
		},
		{
			name:             "no flags, args: 1 file, with multiple lines, ending WITH a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye\n"},
			expectedOutput:   "hello\nfrom the\ntest file\ngoodbye\n",
		},
		{
			name:             "no flags, args: 1 file, with multiple lines, ending WITHOUT a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye"},
			expectedOutput:   "hello\nfrom the\ntest file\ngoodbye",
		},
		{
			name:             "-n flag, args: 1 file, with multiple lines, ending WITH a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye\n"},
			expectedOutput:   "1\thello\n2\tfrom the\n3\ttest file\n4\tgoodbye\n",
		},
		{
			name:             "-n flag, args: 1 file, with multiple lines, ending WITHOUT a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye"},
			expectedOutput:   "1\thello\n2\tfrom the\n3\ttest file\n4\tgoodbye",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// get the original working directory
			originalWorkingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to store original working directory: %v", err)
			}

			// create a temporary directory to create the test files in
			temporaryDirectory, err := os.MkdirTemp(".", "go-cat-temporary-directory")
			if err != nil {
				t.Fatalf("failed to create a temporary directory: %v", err)
			}
			defer os.RemoveAll(temporaryDirectory)

			// create the test files
			for index, testFile := range testCase.testFiles {
				testFilePath := filepath.Join(temporaryDirectory, testFile)
				err := os.WriteFile(testFilePath, []byte(testCase.testFileContents[index]), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
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
					t.Fatalf("failed to change to the temporary directory: %v", err)
				}
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

			actualOutput := string(pipeReadBytes)

			if actualOutput != testCase.expectedOutput {
				t.Errorf("output: actual %v | expected %v", actualOutput, testCase.expectedOutput)
			}
		})
	}
}
