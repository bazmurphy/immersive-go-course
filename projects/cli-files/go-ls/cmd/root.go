package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
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

	// get the non-flag command line arguments
	args := flag.Args()

	// make a slice of paths to iterate over (if more than one argument is passed)
	var paths []string

	if len(args) == 0 {
		// if there are no arguments passed in then use the current working directory
		workingDirectory, _ := os.Getwd()
		paths = append(paths, workingDirectory)
	} else {
		// if there are arguments
		paths = append(paths, args...)
	}

	// loop over the paths
	for index, path := range paths {
		// get the "FileInfo" about the path
		pathInfo, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not read the file/directory information: %v\n", err)
			os.Exit(2)
		}

		// if the path is a file then print it back to the user
		if !pathInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "%v\n", os.Args[1])
			break
		}

		// read from that directory
		directory, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading from the directory: %v\n", err)
			os.Exit(2)
		}

		// if we have more than one path then print
		if len(paths) > 1 {
			fmt.Fprintf(os.Stdout, "%v:\n", path)
		}

		// loop through the files/folders and print them
		for _, file := range directory {
			// ignore any hidden files
			if !strings.HasPrefix(file.Name(), ".") {
				fmt.Fprintf(os.Stdout, "%v\n", file.Name())
			}
		}

		// add a newline to separate multiple paths
		if len(paths) > 1 && index < len(paths)-1 {
			fmt.Fprint(os.Stdout, "\n")
		}
	}
}
