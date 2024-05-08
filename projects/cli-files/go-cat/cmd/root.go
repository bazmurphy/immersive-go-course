package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

var (
	numberFlag = flag.Bool("n", false, "number all output lines")
)

func Execute() {
	flag.Parse()

	// get the non-flag arguments
	args := flag.Args()

	// if there is no non-flag argument given then error
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: no filename provided\n")
		os.Exit(2)
	}

	for _, filename := range args {
		// try to get the file information
		fileInfo, err := os.Stat(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-cat: %s: No such file or directory\n", filename)
			continue
		}

		// if the file is a directory then warn the user
		if fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "go-cat: %s: Is a directory\n", filename)
			continue
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-cat: %v: failed to open the file\n", file)
			continue
		}
		defer file.Close()

		lineNumber := 1

		// -------------------- METHOD 1 bufio.NewScanner

		// bufio.NewScanner() creates a new Scanner to read input from a specified io.Reader
		scanner := bufio.NewScanner(file)

		// Scan() advances the Scanner to the next token (default is a line)
		// and returns true if there are more tokens to read
		// and returns false when it reaches the end of the input or encounters an error
		for scanner.Scan() {
			// Text() returns the most recently scanned token as a string
			// it should be called only after a successful call to Scan()
			line := scanner.Text()

			if *numberFlag {
				fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
				lineNumber++
			} else {
				// need to somehow work out whether this has a newline or not at the end
				// and don't add the \n if it doesn't
				fmt.Fprintf(os.Stdout, "%s", line)
			}
		}

		// Err() returns the first non-EOF error that was encountered by the scanner
		err = scanner.Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-cat: %v: failed to read the file\n", file)
		}

		// -------------------- METHOD 2 bufio.NewReader

		// reader := bufio.NewReader(file)

		// // attempt to read the file
		// for {
		// 	line, err := reader.ReadString('\n')
		// 	if err != nil {
		// 		if err == io.EOF {
		// 			break
		// 		}
		// 		fmt.Fprintf(os.Stderr, "go-cat: %v: failed to read line %d", file, lineNumber)
		// 		continue
		// 	}

		// 	if *numberFlag {
		// 		fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
		// 		lineNumber++
		// 	} else {
		// 		fmt.Fprint(os.Stdout, line)
		// 	}
		// }
	}
}
