package cmd

import (
	"bufio"
	"fmt"
	"os"
)

func Execute() {

	// if there is no argument given then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "no filename provided\n")
		os.Exit(2)
	}

	// if there is an argument provided then save that argument as the filename
	filename := os.Args[1]

	// print out the filename to check
	// fmt.Println("--- DEBUG filename", filename)

	// check if the file exists
	// fileInfo, err := os.Stat(filename)
	_, err := os.Stat(filename)

	// if the file does not exist then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "no such file or directory: %s\n", filename)
		os.Exit(1)
	}

	// print out the fileInfo to check
	// fmt.Println("--- DEBUG fileInfo", fileInfo)

	// attempt to open the file
	file, err := os.Open(filename)

	// if the file cannot be opened then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open the file: %v\n", err)
		os.Exit(1)
	}

	// print out the file to check
	// fmt.Println("--- DEBUG file", file)

	// make sure to close the file at the end of the function
	defer file.Close()

	// make a scanner to read from the file
	scanner := bufio.NewScanner(file)

	// use the scanner to scan through the file, line by line, printing out each line
	for scanner.Scan() {
		fmt.Fprint(os.Stdout, scanner.Text(), "\n")
	}

}
