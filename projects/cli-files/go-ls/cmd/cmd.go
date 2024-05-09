package cmd

import (
	"fmt"
	"os"
	"strings"
)

type Flags struct {
	Help bool
}

func Execute(flags *Flags, args []string) error {
	if flags.Help {
		fmt.Fprintf(os.Stdout, "go-ls help message")
		return nil
	}

	// if there are no arguments passed in then use the current working directory
	if len(args) == 0 {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get the current working directory: %v", err)
		}
		args = append(args, workingDirectory)
	}

	// loop over the args (paths)
	for index, path := range args {
		// get the "FileInfo" about the path
		pathInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("could not read the file/directory information: %v", err)
		}

		// if the path is a file then simply print it back to the user
		if !pathInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "%v\n", path)
			break
		}

		// otherwise read from that directory
		directory, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading from the directory: %v\n", err)
		}

		// if we have more than one arg then first print the "path:"
		if len(args) > 1 {
			fmt.Fprintf(os.Stdout, "%v:\n", path)
		}

		// loop through the files/directories and print them
		for _, file := range directory {
			// ignore any hidden files
			if !strings.HasPrefix(file.Name(), ".") {
				fmt.Fprintf(os.Stdout, "%v\n", file.Name())
			}
		}

		// add a newline to separate multiple paths
		if len(args) > 1 && index < len(args)-1 {
			fmt.Fprint(os.Stdout, "\n")
		}
	}

	return nil
}
