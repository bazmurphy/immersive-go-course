A CLI tool that:

1. reads an input CSV containing URLs
2. downloads images by URL
3. processes them using ImageMagick to make them monochrome/grayscale
4. uploads the results to Amazon AWS S3 cloud storage
5. writes an output CSV describing what it did

It should run like this: go run . --input input.csv --output output.csv

You should modify and extend the main.go file we’ve supplied in this directory, re-using some bits of it.

The tool should (in order of priority):

1. Write thorough logs (log.Println and log.Printf) to describe what it is doing, including errors
2. Validate the input CSV to ensure it only has one column, url
3. Gracefully handle failures & continue to process the input CSV even if one row fails
4. Support a configurable AWS region and S3 bucket via environment variables AWS_REGION and S3_BUCKET

## Reading a CSV

- The built-in encoding/csv package is the one to use to read and write the CSV files.

## Downloading the file

We can use the the standard http package to download the image. Things to watch out for:

- HTTP requests can fail - remember to catch the error!
- HTTP requests can “succeed” but with a non-200 status code. Think about what that could mean!
- How can you make sure the downloaded data is an image, and not some other of file?

## Image processing

- Most of the ImageMagick code, which grayscales the image, is written for you. This shouldn’t change too much.
