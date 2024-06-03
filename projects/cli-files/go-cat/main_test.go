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

// from Daniel Session:

// package main

// import (
// 	"bytes"
// 	"os"
// 	"os/exec"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestSimpleSuccess(t *testing.T) {
// 	temporaryFile, err := os.Create("testfile.txt")
// 	if err != nil {
// 		t.Errorf("error creating the temporary file")
// 	}

// 	expected := "hello world\n"

// 	// TODO: ReadString('\n') is a trap you get the value and the EOF
// 	// so it breaks with no newline -- make a test case for this

// 	// write a string to the temporary file
// 	_, err = temporaryFile.WriteString(expected)
// 	if err != nil {
// 		t.Errorf("error writing a string to the temporary file")
// 	}
// 	// TODO: Talk about sync/flush

// 	// close the file after creating it and writing to it
// 	temporaryFile.Close()

// 	// remove the temporary file after the test
// 	defer os.Remove(temporaryFile.Name())

// 	var capturedOutput bytes.Buffer
// 	var errorOutput bytes.Buffer

// 	///////////////
// 	// Test starts being different
// 	///////////////

// 	// TODO: Build this and find the path to it
// 	cmd := exec.Command("go-cat", temporaryFile.Name())
// 	cmd.Stdout = &capturedOutput
// 	cmd.Stderr = &errorOutput
// 	err = cmd.Run()

// 	require.NoError(t, err)
// 	require.Equal(t, expected, capturedOutput.String())
// }
