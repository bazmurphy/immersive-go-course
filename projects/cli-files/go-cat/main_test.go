// projects/cli-files/go-cat/main_test.go

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		name           string
		flags          []string
		args           []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "no flags, args: 1 file",
			flags:          []string{},
			args:           []string{"assets/sample.txt"},
			expectedStdout: "this is the first line\nthis is line 2\nthis is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\nthis is line 4\nthis is the last line (deliberately with no newline)",
			expectedStderr: "",
		},
		{
			name:           "no flags, no args",
			flags:          []string{},
			args:           []string{},
			expectedStdout: "",
			expectedStderr: "go-cat: no filename provided",
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

			if actualStdout != testCase.expectedStdout {
				t.Errorf("actual: %v | expected: %v", actualStdout, testCase.expectedStdout)
			}

			if actualStderr != testCase.expectedStderr {
				t.Errorf("actual: %v | expected: %v", actualStderr, testCase.expectedStderr)
			}
		})
	}
}

func TestMainOSExec(t *testing.T) {
	testCases := []struct {
		name           string
		flags          []string
		args           []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "arg: 1 file",
			flags:          []string{},
			args:           []string{"assets/sample.txt"},
			expectedStdout: "this is the first line\nthis is line 2\nthis is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\nthis is line 4\nthis is the last line (deliberately with no newline)",
			expectedStderr: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// janky way to add all the args together as a single string:
			cmd := exec.Command("go", append([]string{"run", "."}, testCase.args...)...)

			// create buffers to capture the stdout/stderr
			var stdoutBuffer bytes.Buffer
			var stderrBuffer bytes.Buffer

			// set the stdout/stderr of the command to those buffers^
			cmd.Stdout = &stdoutBuffer
			cmd.Stderr = &stderrBuffer

			// run the command
			err := cmd.Run()
			if err != nil {
				t.Errorf("error running the command: %v", err)
				return
			}

			actualStdout := stdoutBuffer.String()
			actualStderr := stderrBuffer.String()

			if actualStdout != testCase.expectedStdout {
				t.Errorf("actualStdout : actual %v | expected %v", actualStdout, testCase.expectedStdout)
			}

			if actualStderr != testCase.expectedStderr {
				t.Errorf("actualStderr : actual %v | expected %v", actualStderr, testCase.expectedStderr)
			}
		})
	}
}
