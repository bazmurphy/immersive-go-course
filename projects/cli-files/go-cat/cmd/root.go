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

	// check if the file exists
	fileInfo, err := os.Stat(filename)

	if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "%s : Is a directory\n", filename)
		os.Exit(1)
	}

	// if the file does not exist then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "no such file or directory: %s\n", filename)
		os.Exit(2)
	}

	// attempt to read the file
	file, err := os.ReadFile(filename)

	// if we cannot read the file then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the file: %v\n", err)
		os.Exit(2)
	}

	// convert the file into a string
	fileContents := string(file)

	// if the -n flag was provided
	if *numberFlag {
		// split the string into lines using the newline delimiter
		lines := strings.Split(fileContents, "\n")
		// loop over the lines printing them with a line number prefix
		for index, line := range lines {
			fmt.Fprintf(os.Stdout, "%d  %s\n", index+1, line)
		}
		return
	}

	// directly print the file contents
	fmt.Fprint(os.Stdout, fileContents)

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
