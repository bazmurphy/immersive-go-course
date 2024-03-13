package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// if there is no argument given then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "no filename provided\n")
		os.Exit(2)
	}

	// if there is an argument given, get that argument
	filename := os.Args[1]

	// print out the filename to check
	// fmt.Println("--- DEBUG filename", filename)

	// check if the file exists
	// fileInfo, err := os.Stat(filename)
	_, err := os.Stat(filename)

	// if it does not exist then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "file does not exist: %s\n", filename)
		os.Exit(1)
	}

	// print out the fileInfo to check
	// fmt.Println("--- DEBUG fileInfo", fileInfo)

	// check if the file can be opened
	file, err := os.Open(filename)

	// if it cannot be opened then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open the file: %v\n", err)
		os.Exit(1)
	}

	// print out the fileInfo to check
	// fmt.Println("--- DEBUG file", file)

	// make sure to close the file at the end of the function
	defer file.Close()

	// make a scanner and read from the file
	scanner := bufio.NewScanner(file)

	// read through the file, line by line, and print the line
	for scanner.Scan() {
		fmt.Fprint(os.Stdout, scanner.Text(), "\n")
	}
}
