package main

import (
	"bytes"
	"fmt"
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
			name:           "no flags, no args",
			expectedStderr: "go-cat: no filename provided",
		},
		{
			name:           "no flags, args: 1 file",
			args:           []string{"assets/sample1.txt"},
			expectedStdout: "this is the first line\nthis is line 2\nthis is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\nthis is line 4\nthis is the last line (deliberately with no newline)",
		},
		{
			name:           "no flags, args: 2 files",
			args:           []string{"assets/sample1.txt", "assets/sample2.txt"},
			expectedStdout: "this is the first line\nthis is line 2\nthis is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\nthis is line 4\nthis is the last line (deliberately with no newline)this is line 1 from the second sample file\nthis is line 2 from the second sample file",
		},
		{
			name:           "-n flag, args: 1 file",
			flags:          []string{"-n"},
			args:           []string{"assets/sample1.txt"},
			expectedStdout: "1	this is the first line\n2	this is line 2\n3	this is line 3 (deliberately longer to test text wrapping) this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3 this is still line 3\n4	this is line 4\n5	this is the last line (deliberately with no newline)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// construct the command arguments
			commandArgs := []string{"run", "."}
			commandArgs = append(commandArgs, testCase.flags...)
			commandArgs = append(commandArgs, testCase.args...)

			// construct the command
			cmd := exec.Command("go", commandArgs...)

			fmt.Println("DEBUG | cmd", cmd)

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
