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
		expectedStdout   string
		expectedStderr   string
	}{
		{
			name:             "no flags, no args",
			flags:            Flags{},
			args:             []string{},
			testFiles:        []string{},
			testFileContents: []string{},
			expectedStderr:   "",
		},
		{
			name:             "no flags, args: 1 file, with a single line, ending WITH a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\n"},
			expectedStdout:   "hello\n",
		},
		{
			name:             "no flags, args: 1 file, with a single line, ending WITHOUT a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello"},
			expectedStdout:   "hello",
		},
		{
			name:             "-n flag, args: 1 file, with a single line, ending WITH a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\n"},
			expectedStdout:   "1\thello\n",
		},
		{
			name:             "-n flag, args: 1 file, with a single line, ending WITHOUT a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello"},
			expectedStdout:   "1\thello",
		},
		{
			name:             "no flags, args: 1 file, with multiple lines, ending WITH a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye\n"},
			expectedStdout:   "hello\nfrom the\ntest file\ngoodbye\n",
		},
		{
			name:             "no flags, args: 1 file, with multiple lines, ending WITHOUT a new line",
			flags:            Flags{},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye"},
			expectedStdout:   "hello\nfrom the\ntest file\ngoodbye",
		},
		{
			name:             "-n flag, args: 1 file, with multiple lines, ending WITH a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye\n"},
			expectedStdout:   "1\thello\n2\tfrom the\n3\ttest file\n4\tgoodbye\n",
		},
		{
			name:             "-n flag, args: 1 file, with multiple lines, ending WITHOUT a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file.txt"},
			testFiles:        []string{"test-file.txt"},
			testFileContents: []string{"hello\nfrom the\ntest file\ngoodbye"},
			expectedStdout:   "1\thello\n2\tfrom the\n3\ttest file\n4\tgoodbye",
		},
		{
			name:             "no flags, args: 3 files, each with one line, ending WITH a new line",
			flags:            Flags{},
			args:             []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFiles:        []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFileContents: []string{"hello from 1\n", "hello from 2\n", "hello from 3\n"},
			expectedStdout:   "hello from 1\nhello from 2\nhello from 3\n",
		},
		{
			name:             "no flags, args: 3 files, each with one line, ending WITHOUT a new line",
			flags:            Flags{},
			args:             []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFiles:        []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFileContents: []string{"hello from 1", "hello from 2", "hello from 3"},
			expectedStdout:   "hello from 1hello from 2hello from 3",
		},
		{
			name:             "-n flag, args: 3 files, each with one line, ending WITH a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFiles:        []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFileContents: []string{"hello from 1\n", "hello from 2\n", "hello from 3\n"},
			expectedStdout:   "1\thello from 1\n1\thello from 2\n1\thello from 3\n",
		},
		{
			name:             "-n flag, args: 3 files, each with one line, ending WITHOUT a new line",
			flags:            Flags{Number: true},
			args:             []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFiles:        []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
			testFileContents: []string{"hello from 1", "hello from 2", "hello from 3"},
			expectedStdout:   "1\thello from 11\thello from 21\thello from 3",
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
