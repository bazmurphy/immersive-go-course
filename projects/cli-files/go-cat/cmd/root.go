package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func Execute() {
	// define a flag of type boolean with the specific properties (name, value, usage)
	numberFlag := flag.Bool("n", false, "number all output lines")

	// parse the command line flags
	flag.Parse()

	// get the non-flag arguments
	args := flag.Args()
	// fmt.Println("--- DEBUG args", args)

	// if there is no non-flag argument given then error
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: no filename provided\n")
		os.Exit(2)
	}

	// (!) this is very hardcoded... what about more than one file passed to go-cat(?)
	// (!) what about non filename arguments(?)
	filename := args[0]

	// try to get the file information
	fileInfo, err := os.Stat(filename)
	// if we cannot get the file information then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read the file/directory information: %s\n", filename)
		os.Exit(2)
	}

	// if the file is a directory then warn the user
	if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s : Is a directory\n", filename)
		os.Exit(1)
	}

	// open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not open the file: %v\n", err)
		os.Exit(2)
	}
	defer file.Close()

	// use a buffered reader to read from the file
	reader := bufio.NewReader(file)

	lineNumber := 1

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "error: reading line from the file: %v\n", err)
			os.Exit(2)
		}

		if *numberFlag {
			fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
			lineNumber++
		} else {
			fmt.Fprint(os.Stdout, line)
		}
	}
}
