package converter

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

func ConvertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory string, outputCSVRows [][]string) error {
	log.Println("🔵 attempting: to convert images to grayscale...")

	var imagesConvertedCount int

	imagick.Initialize()
	defer imagick.Terminate()

	c := &Converter{
		cmd: imagick.ConvertImageCommand,
	}

	temporaryDownloadsFiles, err := os.ReadDir(temporaryDownloadsDirectory)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to read files from the temporary downloads directory: %v", err)
	}

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

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 4] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], outputFilepath)

		log.Printf("🔵 attempting: to convert %q to %q\n", inputFilepath, outputFilepath)

		err := c.Grayscale(inputFilepath, outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to convert the image to grayscale: %v\n", err)
			continue
		}

		log.Printf("🟢 success: converted %q to %q\n", inputFilepath, outputFilepath)

		imagesConvertedCount++
	}

	log.Printf("🟢 success: converted %d images to grayscale\n", imagesConvertedCount)

	return nil
}
