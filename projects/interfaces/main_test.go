package main

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestOurBytesBuffer(t *testing.T) {
	t.Run("returns the same bytes it was created with", func(t *testing.T) {
		// create a new buffer instance with some initial bytes
		ourBuffer := NewOurBuffer([]byte{1, 2, 3})

		// get the bytes from the buffer using the Bytes() method
		got := ourBuffer.Bytes()

		expectedByteSlice := []byte{1, 2, 3}

		// error if the bytes from the buffer and the expected bytes are not the same
		if !bytes.Equal(got, expectedByteSlice) {
			t.Errorf("error byte slices are not equal: got %q | expected %q", got, expectedByteSlice)
		}
	})

	t.Run("returns both initial bytes and the extra bytes", func(t *testing.T) {
		// create a new buffer instance with some initial bytes
		ourBuffer := NewOurBuffer([]byte{1, 2, 3})

		// write the extra bytes to the buffer
		numberOfBytesWritten, err := ourBuffer.Write([]byte{4, 5, 6})

		if err != nil {
			t.Errorf("error received but expected none: got %q | expected %v", err, nil)
		}

		expectedByteSlice := []byte{1, 2, 3, 4, 5, 6}
		bytesWritten := []byte{4, 5, 6}
		expectedNumberOfBytesWritten := len(bytesWritten)

		// get the bytes from the buffer using the Bytes() method
		got := ourBuffer.Bytes()

		// error if the bytes from the buffer and the expected bytes are not the same
		if !bytes.Equal(got, expectedByteSlice) {
			t.Errorf("error byte slices are not equal: got %q | expected %q", got, expectedByteSlice)
		}

		if numberOfBytesWritten != expectedNumberOfBytesWritten {
			t.Errorf("error number of bytes written: got %d | expected %d", numberOfBytesWritten, expectedNumberOfBytesWritten)
		}
	})

	t.Run("slice big enough to read all of the bytes in the buffer", func(t *testing.T) {
		// create a new buffer instance with some initial bytes
		ourBuffer := NewOurBuffer([]byte{1, 2, 3})

		// make a byte slice with a size of the buffer length
		bytesStoreSlice := make([]byte, ourBuffer.Len())

		// read from the buffer
		numberOfBytesRead, err := ourBuffer.Read(bytesStoreSlice)

		initialBytes := []byte{1, 2, 3}

		if err != nil {
			t.Errorf("error received but expected none: got %q | expected %v", err, nil)
		}

		// error if the byte slices are not the same
		if !bytes.Equal(bytesStoreSlice, initialBytes) {
			t.Errorf("byte slices are not equal: got %v | want %v", bytesStoreSlice, initialBytes)
		}

		// error if the number of bytes read is not the same as the length of the initial bytes
		if numberOfBytesRead != len(initialBytes) {
			t.Errorf("number of bytes read: got %d | want %d", numberOfBytesRead, len(initialBytes))
		}
	})

	t.Run("slice smaller than all of the bytes in the buffer", func(t *testing.T) {
		// create a new buffer instance with some initial bytes
		ourBuffer := NewOurBuffer([]byte{1, 2, 3, 4, 5, 6, 7})

		// initialise some slices to keep track of:
		var byteReadValueHistory [][]byte // the byte values read from the buffer each read
		var byteReadNumberHistory []int   // the number of values read from the buffer each read
		var bytesRemaining []byte         // the remaining bytes after the read fails

		// while the buffer is not empty, read from it
		for {
			// deliberate use a slice smaller than the length of bytes in the buffer
			bytesStoreSliceSmall := make([]byte, 2)

			// read the buffer into the small slice
			// (!) if there is not enough values to read into slice, the rest will remain nil values (from make) in this case 0
			numberOfBytesRead, err := ourBuffer.Read(bytesStoreSliceSmall)

			//(!) EOF is the error returned by Read when no more input is available
			// if no more input is available then store the rest in bytesRemaining and exit the loop
			if err == io.EOF {
				// b.Bytes() gives us the unread portion of the buffer
				bytesRemaining = ourBuffer.Bytes()
				break
			}

			if err != nil {
				t.Errorf("error received but expected none: got %q | expected %v", err, nil)
			}

			// append the values to the history slices
			byteReadValueHistory = append(byteReadValueHistory, append([]byte{}, bytesStoreSliceSmall...))
			byteReadNumberHistory = append(byteReadNumberHistory, numberOfBytesRead)
		}

		// establish the expected values
		expectedByteReadValueHistory := [][]byte{{1, 2}, {3, 4}, {5, 6}, {7, 0}}
		expectedByteReadNumberHistory := []int{2, 2, 2, 1}
		expectedBytesRemaining := []byte{}

		// reflect is resource intensive(?)... how else can i make object comparisons on pass by reference types: struct, slice, map, etc)
		if !reflect.DeepEqual(byteReadNumberHistory, expectedByteReadNumberHistory) {
			t.Errorf("byteReadNumberHistory: got %v | want %v", byteReadNumberHistory, expectedByteReadNumberHistory)
		}

		if !reflect.DeepEqual(byteReadValueHistory, expectedByteReadValueHistory) {
			t.Errorf("byteReadValueHistory: got %v | want %v", byteReadValueHistory, expectedByteReadValueHistory)
		}

		if !reflect.DeepEqual(bytesRemaining, expectedBytesRemaining) {
			t.Errorf("bytesRemaining: got %v | want %v", bytesRemaining, expectedBytesRemaining)
		}
	})
}

func TestFilteringPipe(t *testing.T) {
	type TestCase struct {
		name     string
		input    string
		expected string
	}

	testCases := []TestCase{
		{name: "digits after equals", input: "start=1, end=10", expected: "start=, end="},
		{name: "digits after every word", input: "hello123 and456 goodbye789", expected: "hello and goodbye"},
		{name: "digits wrapping a word", input: "010101binary010101", expected: "binary"},
		{name: "digit/character pattern", input: "1x2y3z", expected: "xyz"},
		{name: "all digits", input: "010101", expected: ""},
		{name: "no digits", input: "abcdef", expected: "abcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create a new bytes.Buffer (this implements the io.Writer interface)
			someWriter := &bytes.Buffer{}

			// create a new NewFilterPipe instance passing it the bytes.Buffer (as the io.Writer(?))
			filteringPipe := NewFilteringPipe(someWriter)

			// write the input to the filter pipe
			filteringPipe.Write([]byte(tc.input))

			// get the output string from the buffer
			got := someWriter.String()

			// error if the output (got) doesn't match the expected output (want)
			if got != tc.expected {
				t.Errorf("got %v | expected %v", got, tc.expected)
			}
		})
	}
}