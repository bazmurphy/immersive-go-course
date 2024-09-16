package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type ConvertImageCommand func(args []string) (*imagick.ImageCommandResult, error)

type Converter struct {
	cmd ConvertImageCommand
}

func (c *Converter) Grayscale(inputFilepath string, outputFilepath string) error {
	// Convert the image to grayscale using imagemagick
	// We are directly calling the convert command
	_, err := c.cmd([]string{
		"convert", inputFilepath, "-set", "colorspace", "Gray", outputFilepath,
	})
	return err
}

type ConvertedImageObject struct {
	ImageFilepath string
	ID            int
}

func ConvertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory string) ([]ConvertedImageObject, error) {
	log.Println("🔵 attempting: to convert images to grayscale...")

	temporaryDownloadsFiles, err := os.ReadDir(temporaryDownloadsDirectory)
	if err != nil {
		return nil, fmt.Errorf("🔴 error: failed to read files from the temporary downloads directory: %v", err)
	}

	imagick.Initialize()
	defer imagick.Terminate()

	c := &Converter{
		cmd: imagick.ConvertImageCommand,
	}

	var convertedImageObjects []ConvertedImageObject

	for index, file := range temporaryDownloadsFiles {
		// ignore directories
		if file.IsDir() {
			log.Printf("🟠 warn: ignoring a directory...\n")
			continue
		}

		inputFilepath := filepath.Join(temporaryDownloadsDirectory, file.Name())

		fileExtension := filepath.Ext(file.Name())

		fileName := file.Name()[:len(file.Name())-len(fileExtension)]

		outputFilename := fmt.Sprintf("%s-grayscale%s", fileName, fileExtension)

		outputFilepath := filepath.Join(temporaryGrayscaleDirectory, outputFilename)

		log.Printf("🔵 attempting: to convert %q to %q\n", inputFilepath, outputFilepath)

		err := c.Grayscale(inputFilepath, outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to convert the image to grayscale: %v\n", err)
			continue
		}

		log.Printf("🟢 success: converted %q to %q\n", inputFilepath, outputFilepath)

		convertedImageObject := ConvertedImageObject{
			ImageFilepath: outputFilepath,
			ID:            index + 1,
		}

		convertedImageObjects = append(convertedImageObjects, convertedImageObject)
	}

	log.Printf("🟢 success: converted %d images to grayscale\n", len(convertedImageObjects))

	return convertedImageObjects, nil
}
