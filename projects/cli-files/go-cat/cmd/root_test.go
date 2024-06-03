package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExec(t *testing.T) {
	t.Run("go-cat on a filename with no flags", func(t *testing.T) {
		// make a temporary file (deliberately not using os.CreateTemp)
		temporaryFile, err := os.Create("testfile.txt")
		if err != nil {
			t.Errorf("error creating the temporary file")
		}

		expected := "hello world\n"

		// TODO: ReadString('\n') is a trap you get the value and the EOF
		// so it breaks with no newline -- make a test case for this

		// write a string to the temporary file
		_, err = temporaryFile.WriteString(expected)
		if err != nil {
			t.Errorf("error writing a string to the temporary file")
		}

		// TODO: Talk about sync/flush

		// close the file after creating it and writing to it
		temporaryFile.Close()

		// remove the temporary file after the test
		defer os.Remove(temporaryFile.Name())

		// Assert that:
		//  * Execute succeeded
		//  * Output of execute was "hello world".

		var capturedOutput bytes.Buffer
		var errorOutput bytes.Buffer

		actualExitCode, err := Execute(temporaryFile.Name(), false, &capturedOutput, &errorOutput)
		expectedExitCode := 0

		require.NoError(t, err)

		require.Equal(t, expectedExitCode, actualExitCode)

		require.Equal(t, expected, capturedOutput.String())
	})

	// t.Run("go-cat on a filename with the -n flag", func(t *testing.T) {
	// 	// to do
	// 	Execute()
	// })

	// t.Run("go-cat on no filename", func(t *testing.T) {
	// 	// to do
	// 	Execute()
	// })
}
