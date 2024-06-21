package main

import (
	"io"
	"unicode"
)

// ---------- Implementing our own bytes.Buffer

// the original Buffer type definition:
//
//	type Buffer struct {
//		buf      []byte // contents are the bytes buf[off : len(buf)]
//		off      int    // read at &buf[off], write at &buf[len(buf)]
//		lastRead readOp // last read operation, so that Unread* can work correctly.
//	}
type OurByteBuffer struct {
	buf []byte
}

// the original Buffer Bytes() method signature:
// func (b *Buffer) Bytes() []byte
func (b *OurByteBuffer) Bytes() []byte {
	// simply return the buffer
	return b.buf
}

// the original Buffer Write() method signature:
// func (b *Buffer) Write(p []byte) (n int, err error)
func (b *OurByteBuffer) Write(p []byte) (int, error) {
	// Q: Why is no error returned from append(?)
	// A: The append function in Go doesn't generally return an error because it performs a safe operation that cannot fail in most cases.
	// The reason for this is related to how Go handles memory allocation and slice operations.

	// append the byte slice values (passed to the function) to the buffer
	b.buf = append(b.buf, p...)

	// return the number of bytes that were appended, and an error... although it can never be not nil(??)
	return len(p), nil
}

// the original Buffer Read() method signature:
// func (b *Buffer) Read(p []byte) (n int, err error)
func (b *OurByteBuffer) Read(p []byte) (int, error) {
	// if the buffer is empty then return 0 and io.EOF
	if len(b.buf) == 0 {
		// Q: What is is.EOF ?
		// A: EOF stands for End of File
		// When a function returns io.EOF, it means that there's no more data to be read from the input source.
		return 0, io.EOF
	}

	// copy the buffer data into the provided byte slice
	numberOfElementsCopied := copy(p, b.buf)

	// update the buffer by removing the elements that were read
	// (making a slice starting at the index of how many elements were read)
	b.buf = b.buf[numberOfElementsCopied:]

	// return the number of bytes read and nil error
	return numberOfElementsCopied, nil
}

// the original Buffer Len() method signature:
// func (b *Buffer) Len() int
func (b *OurByteBuffer) Len() int {
	// return the length of the buffer
	return len(b.buf)
}

// the original bytes.NewBuffer() function signature:
// func NewBuffer(buf []byte) *Buffer
func NewOurBuffer(initialByteSlice []byte) *OurByteBuffer {
	// create a new OurByteBuffer instance
	// and populate the buffer with the initialByteSlice
	// returning a pointer to the new OurByteBuffer instance
	return &OurByteBuffer{buf: initialByteSlice}
}

// ---------- Implementing a custom filter

// define a new struct
// it's only field is a writer of type io.Writer
type FilteringPipe struct {
	writer io.Writer
}

// the original io.Writer Write() method
// Write(p []byte) (n int, err error)
func (fp *FilteringPipe) Write(p []byte) (int, error) {
	// create a new slice to store the filtered bytes in
	// var filteredBytesSlice []byte

	// // loop through the slice of bytes (that was passed in as an argument)
	// for _, byte := range p {
	// 	// if the byte is not a digit
	// 	if !unicode.IsDigit(rune(byte)) {
	// 		// then append it to the slice
	// 		filteredBytesSlice = append(filteredBytesSlice, byte)
	// 	}
	// }

	// check the filteredBytesSlice
	// fmt.Println(string(filteredBytesSlice))

	// but how to return the filteredBytesSlice correctly?

	// Searched and found this.. but I _REALLY_ don't understand it...
	// return fp.writer.Write(filteredBytesSlice)

	// This line calls the Write() method of the underlying writer (writer) that's embedded within the FilteringPipe struct.
	// It passes the filteredBytesSlice to this writer, essentially sending the filtered data to the next stage in the pipeline.
	// This allows you to chain multiple filters or write the filtered data to a file, network socket, or any other destination that supports writing.

	// The Write() method of the FilteringPipe does not directly return the filtered bytes themselves.
	// Instead, it returns the values returned by the Write() method of the embedded writer.

	// The Write() method filters the input data, removes digits, and forwards the filtered bytes to the underlying writer.
	// It propagates the return values from the underlying writer to provide information about the write operation's success or failure.
	// This mechanism enables you to chain multiple filters or write the filtered data to any destination that implements the io.Writer interface.

	// By returning the results of the underlying writer's Write() method, the FilteringPipe maintains consistency with the io.Writer interface and enables seamless integration with other parts of the Go I/O system.

	// second implementation ----------
	var totalBytesWritten int
	var writeError error

	// loop through the individual bytes in the slice of bytes passed in as argument p
	for _, individualByte := range p {
		if !unicode.IsDigit(rune(individualByte)) {
			// use the underlying Write() method to write the individual byte to (what??)
			numberOfBytesWritten, err := fp.writer.Write([]byte{individualByte})
			totalBytesWritten += numberOfBytesWritten
			if err != nil {
				writeError = err
				break
			}
		}
	}

	return totalBytesWritten, writeError
}

// constructor for FilteringPipe, it takes in an io.Writer as an argument
func NewFilteringPipe(w io.Writer) *FilteringPipe {
	return &FilteringPipe{writer: w}
}
