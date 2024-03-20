package main

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

// ---------- Testing OurBytesBuffer

func TestOurBytesBuffer(t *testing.T) {
	t.Run("returns the same bytes it was created with", func(t *testing.T) {
		// create a byte slice with bytes
		want := []byte{1, 2, 3}

		// create a new buffer instance with the bytes
		// b := bytes.NewBuffer(want)
		b := NewOurBuffer(want)

		// get the bytes from the buffer using the Bytes() method
		got := b.Bytes()

		// error if the bytes from the buffer and the expected bytes are not the same
		if !bytes.Equal(got, want) {
			t.Errorf("got %q | want %q", got, want)
		}
	})

	t.Run("returns both initial bytes and the extra bytes", func(t *testing.T) {
		// create a byte slice with initial bytes
		initialBytes := []byte{1, 2, 3}

		// create a new buffer instance with the initial bytes
		// b := bytes.NewBuffer(initialBytes)
		b := NewOurBuffer(initialBytes)

		// create a byte slice with extra bytes
		extraBytes := []byte{4, 5, 6}

		// write the extra bytes to the buffer
		_, err := b.Write(extraBytes)

		// (?) necessary to prevent the test panicking
		if err != nil {
			t.Errorf("error writing to the buffer: %q", err)
		}

		// get the bytes from the buffer using the Bytes() method
		got := b.Bytes()

		// create a byte slice with the expected bytes
		want := []byte{1, 2, 3, 4, 5, 6}

		// error if the bytes from the buffer and the expected bytes are not the same
		if !bytes.Equal(got, want) {
			t.Errorf("bytes are not equal: got %q | want %q", got, want)
		}

		// error if the lengths of the bytes from the buffer and the expected bytes are not the same
		if len(got) != len(want) {
			t.Errorf("bytes slice lengths are not the same: got %q | want %q", len(got), len(want))
		}
	})

	t.Run("slice big enough to read all of the bytes in the buffer", func(t *testing.T) {
		// create a byte slice with initial bytes
		initialBytes := []byte{1, 2, 3}

		// create a new buffer instance with the initial bytes
		// b := bytes.NewBuffer(initialBytes)
		b := NewOurBuffer(initialBytes)

		// make a byte slice with a size of the buffer length
		bytesStoreSlice := make([]byte, b.Len())

		// read from the buffer
		numberOfBytesRead, err := b.Read(bytesStoreSlice)

		// (?) necessary to prevent the test panicking
		if err != nil {
			t.Errorf("error reading from the buffer: %q", err)
		}

		// error if the number of bytes read is not the same as the length of the initial bytes
		if numberOfBytesRead != len(initialBytes) {
			t.Errorf("number of bytes read: got %d | want %d", numberOfBytesRead, len(initialBytes))
		}

		// error if the byte slices are not the same
		if !bytes.Equal(bytesStoreSlice, initialBytes) {
			t.Errorf("byte slices are not equal: got %v | want %v", bytesStoreSlice, initialBytes)
		}
	})

	t.Run("slice smaller than all of the bytes in the buffer", func(t *testing.T) {
		// create a byte slice with initial bytes
		initialBytes := []byte{1, 2, 3, 4, 5, 6, 7}

		// create a new buffer instance with the initial bytes
		// b := bytes.NewBuffer(initialBytes)
		b := NewOurBuffer(initialBytes)

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
			numberOfBytesRead, err := b.Read(bytesStoreSliceSmall)

			//(!) EOF is the error returned by Read when no more input is available
			// if no more input is available then store the rest in bytesRemaining and exit the loop
			if err == io.EOF {
				// b.Bytes() gives us the unread portion of the buffer
				bytesRemaining = b.Bytes()
				break
			}

			// (?) necessary to prevent the test panicking
			if err != nil {
				t.Errorf("error reading from the buffer: %q", err)
				break
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
	// define a struct to hold the test cases
	type TestCase struct {
		name  string
		input string
		want  string
	}

	// create a slice of test cases (of type TestCase)
	testCases := []TestCase{
		{name: "digits after equals", input: "start=1, end=10", want: "start=, end="},
		{name: "digits after every word", input: "hello123 and456 goodbye789", want: "hello and goodbye"},
		{name: "digits wrapping a word", input: "010101binary010101", want: "binary"},
		{name: "digit/character pattern", input: "1x2y3z", want: "xyz"},
		{name: "all digits", input: "010101", want: ""},
		{name: "no digits", input: "abcdef", want: "abcdef"},
	}

	// loop over the test cases
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// create a new bytes.Buffer (this implements the io.Writer interface)
			someWriter := &bytes.Buffer{}

			// create a new NewFilterPipe instance passing it the bytes.Buffer (as the io.Writer(?))
			filteringPipe := NewFilteringPipe(someWriter)

			// write the input to the filter pipe
			filteringPipe.Write([]byte(testCase.input))

			// get the output string from the buffer
			got := someWriter.String()

			// if the output (got) doesn't match the expected output (want)
			if got != testCase.want {
				t.Errorf("got %v | want %v", got, testCase.want)
			}
		})
	}
}
