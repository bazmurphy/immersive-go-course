// projects/cli-files/go-cat/main_test.go

package main

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		name              string
		flags             []string
		args              []string
		expectedStdOutput string
		expectedStdErr    string
	}{
		{
			name:              "no flags, args: 1 file",
			flags:             []string{},
			args:              []string{"assets/sample.txt"},
			expectedStdOutput: "this is the first line\nthis is line 2\nthis is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\nthis is line 4\nthis is the last line (deliberately with no newline)",
			expectedStdErr:    "",
		},
		{
			name:              "no flags, no args",
			flags:             []string{},
			args:              []string{},
			expectedStdOutput: "",
			expectedStdErr:    "go-cat: no filename provided",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// at the start of the test: store the original args
			originalArgs := os.Args

			// if there are any arguments then add them
			if len(testCase.args) > 0 {
				os.Args = append(os.Args, testCase.args...)
			}

			// store the original stdout/stderr
			originalStdout := os.Stdout
			originalStderr := os.Stderr

			// create a pipe to capture the output
			stdoutPipeRead, stdoutPipeWrite, _ := os.Pipe()
			stderrPipeRead, stderrPipeWrite, _ := os.Pipe()

			// redirect stdout/stderr to the write end of the pipes
			os.Stdout = stdoutPipeWrite
			os.Stderr = stderrPipeWrite

			main()

			// close the write end of the pipes
			stdoutPipeWrite.Close()
			stderrPipeWrite.Close()

			// read the captured output from the read end of the pipes
			stdoutPipeReadBytes, _ := io.ReadAll(stdoutPipeRead)
			stderrPipeReadBytes, _ := io.ReadAll(stderrPipeRead)
			// and this is bad because i am reading the whole thing at once and not streaming it...
			// which totally defeats the point of reading it line by line earlier...

			actualStdout := string(stdoutPipeReadBytes)
			actualStderr := string(stderrPipeReadBytes)

			// at the end of the test: restore the original stdout/stderr
			os.Stdout = originalStdout
			os.Stderr = originalStderr

			fmt.Println("actualStdout", actualStdout)
			fmt.Println("actualStderr", actualStderr)

			// at the end of the test: restore the original arguments
			os.Args = originalArgs

			if actualStdout != testCase.expectedStdOutput {
				t.Errorf("actual: %v | expected: %v", actualStdout, testCase.expectedStdOutput)
			}

			if actualStderr != testCase.expectedStdErr {
				t.Errorf("actual: %v | expected: %v", actualStderr, testCase.expectedStdErr)
			}
		})
	}
}
