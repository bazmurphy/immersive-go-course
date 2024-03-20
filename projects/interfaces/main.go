package main

import "io"

// the original Buffer type definition:
// type Buffer struct {
// 	buf      []byte // contents are the bytes buf[off : len(buf)]
// 	off      int    // read at &buf[off], write at &buf[len(buf)]
// 	lastRead readOp // last read operation, so that Unread* can work correctly.
// }
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
		// EOF stands for End of File
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
