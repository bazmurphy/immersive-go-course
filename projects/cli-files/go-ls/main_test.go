// projects/cli-files/go-ls/main_test.go

package main

import (
	"flag"
	"io"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		name           string
		flags          []string
		args           []string
		expectedOutput string
	}{
		{
			name:           "no flags, no args",
			flags:          []string{},
			args:           []string{},
			expectedOutput: "assets\ncmd\ngo.mod\nmain.go\nmain_test.go\n",
		},
		{
			name:           "-h flag, no args",
			flags:          []string{"h"},
			args:           []string{},
			expectedOutput: "go-ls help message\n",
		},
		{
			name:           "no flags, . arg",
			flags:          []string{},
			args:           []string{"."},
			expectedOutput: "assets\ncmd\ngo.mod\nmain.go\nmain_test.go\n",
		},
		{
			name:           "no flags, assets arg",
			flags:          []string{},
			args:           []string{"assets"},
			expectedOutput: "dew.txt\nfor_you.txt\nrain.txt\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// (!) i had a load of problems related to flags, the -h flag gets 'set' and then was never 'unset'

			// at the start of the test: reset all the flags
			flag.VisitAll(func(f *flag.Flag) {
				f.Value.Set(f.DefValue)
			})

			// at the start of the test: store the original args
			originalArgs := os.Args

			// if there are any flags then set them
			if len(testCase.flags) > 0 {
				for _, testCaseFlag := range testCase.flags {
					err := flag.Set(testCaseFlag, "true") // this is hard coded and disgusting (they are not all booleans)
					if err != nil {
						t.Fatalf("failed to set the flag %s : %v", testCaseFlag, err)
					}
				}
			}

			// if there are any arguments then add them
			if len(testCase.args) > 0 {
				os.Args = append(os.Args, testCase.args...)
			}

			// store the original stdout
			originalStdout := os.Stdout

			// create a pipe to capture the output
			pipeRead, pipeWrite, _ := os.Pipe()

			// redirect stdout to the write end of the pipe
			os.Stdout = pipeWrite

			main()

			// close the write end of the pipe
			pipeWrite.Close()

			// read the captured output from the read end of the pipe
			pipeReadBytes, _ := io.ReadAll(pipeRead)
			// and this is bad because i am reading the whole thing at once and not streaming it...
			// which totally defeats the point of reading it line by line earlier...

			actualOutput := string(pipeReadBytes)

			// at the end of the test: restore the original stdout
			os.Stdout = originalStdout

			// at the end of the test: restore the original arguments
			os.Args = originalArgs

			if actualOutput != testCase.expectedOutput {
				t.Errorf("actual: %v | expected: %v", actualOutput, testCase.expectedOutput)
			}
		})
	}
}
