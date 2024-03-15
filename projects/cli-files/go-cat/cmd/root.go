package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func Execute() {
	// define a flag of type boolean with the specific properties (name, value, usage)
	numberFlag := flag.Bool("n", false, "number all output lines")

	// parse the command line flags
	flag.Parse()

	// if there is no argument given then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "no filename provided\n")
		os.Exit(2)
	}

	// create a variable to store the file name
	var filename string

	if *numberFlag {
		// if the -n flag is provided the filename will be the third value in the os.Args slice
		filename = os.Args[2]
	} else {
		// if no flag is provided the filename will be the second value in the os.Args slice
		filename = os.Args[1]
	}

	// try to get the file information
	fileInfo, err := os.Stat(filename)

	// if we cannot get the file information then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read the file/directory information: %s\n", filename)
		os.Exit(2)
	}

	// if the file is a directory then warn the user
	if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "%s : Is a directory\n", filename)
		os.Exit(1)
	}

	// try to read the file
	fileContents, err := os.ReadFile(filename)

	// if we cannot read the file then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the file: %v\n", err)
		os.Exit(2)
	}

	// convert the file contents to a string
	fileContentsAsString := string(fileContents)

	// if the -n flag was provided
	if *numberFlag {
		// split the string into lines using the newline delimiter
		lines := strings.Split(fileContentsAsString, "\n")
		// loop over the lines printing them with a line number prefix
		for index, line := range lines {
			fmt.Fprintf(os.Stdout, "%d  %s\n", index+1, line)
		}
		return
	}

	// otherwise directly print the file contents
	fmt.Fprint(os.Stdout, fileContentsAsString)

	// ----------

	// bufio scanner method (currently left here for learning)

	// // attempt to open the file
	// file, err := os.Open(filename)

	// // if the file cannot be opened then error
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "cannot open the file: %v\n", err)
	// 	os.Exit(1)
	// }

	// // make sure to close the file at the end of the function
	// defer file.Close()

	// // make a scanner to read from the file
	// scanner := bufio.NewScanner(file)

	// // use the scanner to scan through the file, line by line, printing out each line

	// // if there is a number flag then prefix it with a line number followed by two spaces
	// if *numberFlag {
	// 	for lineNumber := 1; scanner.Scan(); lineNumber++ {
	// 		fmt.Fprint(os.Stdout, lineNumber, "  ", scanner.Text(), "\n")
	// 	}
	// 	return
	// }

	// for scanner.Scan() {
	// 	fmt.Fprint(os.Stdout, scanner.Text(), "\n")
	// }

}
