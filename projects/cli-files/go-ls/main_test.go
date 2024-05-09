package main

import (
	"bytes"
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
			flags:          []string{},
			args:           []string{},
			expectedStdout: "assets\ncmd\ngo.mod\nmain.go\nmain_test.go\n",
			expectedStderr: "",
		},
		{
			name:           "no flags, arg: 1 folder (with 3 files inside)",
			flags:          []string{},
			args:           []string{"assets"},
			expectedStdout: "dew.txt\nfor_you.txt\nrain.txt\n",
			expectedStderr: "",
		},
		{
			name:           "-h, no args",
			flags:          []string{"-h"},
			args:           []string{""},
			expectedStdout: "go-ls help message",
			expectedStderr: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// construct the command arguments
			commandArgs := []string{"run", "."}
			commandArgs = append(commandArgs, testCase.flags...)
			commandArgs = append(commandArgs, testCase.args...)

			// execute the command
			cmd := exec.Command("go", commandArgs...)

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
