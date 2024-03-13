package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func Execute() {
	// define a flag of type boolean with the specific properties (name, value, usage)
	helpFlag := flag.Bool("h", false, "show help")

	// parse the command line flags
	flag.Parse()

	// print out the value of helpFlag to check
	// (!) deference the pointer with *
	// fmt.Println("--- DEBUG helpFlag", *helpFlag)

	// if the "-h" flag is provided
	if *helpFlag {
		fmt.Fprintf(os.Stdout, "go-ls help message\n")
		return
	}

	// initialise/default the directory path to where go-ls was called from
	directoryPath := "."

	// if go-ls was called with an argument use that as the directory path instead
	if len(os.Args) > 1 {
		directoryPath = os.Args[1]
	}

	// print out the directory path to check
	// fmt.Println("--- DEBUG directoryPath", directoryPath)

	// get the absolute path of that directory
	absolutePath, err := filepath.Abs(directoryPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "absolute path not found: %v\n", err)
		os.Exit(2)
	}

	// print out the absolute path to check
	// fmt.Println("--- DEBUG absolutePath", absolutePath)

	// read from that directory
	directory, err := os.ReadDir(absolutePath)

	// if that directory doesn't exist
	if err != nil {
		fmt.Fprintf(os.Stderr, "no such file or directory: %v\n", err)
		os.Exit(2)
	}

	// print out the directory to check
	// fmt.Println("--- DEBUG directory", directory)

	// loop over the files/directories in the "directory" slice
	for _, file := range directory {
		// fmt.Println("--- DEBUG file", file)
		fmt.Fprintf(os.Stdout, "%v\n", file.Name())
	}
}
