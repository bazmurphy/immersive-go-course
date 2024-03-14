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

	// if the "-h" flag is provided
	if *helpFlag {
		fmt.Fprintf(os.Stdout, "go-ls help message\n")
		return
	}

	// initialise/default the directory path to where go-ls was called from
	directoryPath, err := os.Getwd()

	// if we can't get the current working directory
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get the current working directory: %v\n", err)
		os.Exit(2)
	}

	// if go-ls was called with an argument use that as the directory path instead
	if len(os.Args) > 1 {
		directoryPath = os.Args[1]
	}

	// get the absolute path of that directory
	absolutePath, err := filepath.Abs(directoryPath)

	// if we can't get the absolute path
	if err != nil {
		fmt.Fprintf(os.Stderr, "absolute path not found: %v\n", err)
		os.Exit(2)
	}

	// get the "file" info
	// to be able to establish if the absolute path is a file or directory
	pathInfo, err := os.Stat(absolutePath)

	// if we can't get the "file" info from the absolute path then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read the file/directory information: %v\n", err)
		os.Exit(2)
	}

	// if the path is not a directory (then it must be a file(?)) so print it back to the user
	if !pathInfo.IsDir() {
		fmt.Fprintf(os.Stdout, "%v\n", os.Args[1])
		return
	}

	// read from that directory
	directory, err := os.ReadDir(absolutePath)

	// if there is an error reading from the directory
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading from the directory: %v\n", err)
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
