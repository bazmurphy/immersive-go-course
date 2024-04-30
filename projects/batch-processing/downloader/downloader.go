package downloader

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func DownloadImages(imageUrls []string, temporaryDownloadsDirectory string, outputCSVRows [][]string) error {
	log.Println("🔵 attempting: to download the images from the image urls...")

	var imageDownloadCount int

	for index, imageUrl := range imageUrls {
		// TODO: use context with timeout here (otherwise it can hang infinitely)
		// TODO: use some retry logic here (try 3 times and then give up)
		response, err := http.Get(imageUrl)
		if err != nil {
			log.Printf("🟠 warn: failed to get image url response from url %s\n", imageUrl)
			// TODO: do i want to continue here? as in just move onto the next imageUrl.. no I should retry
			continue
		}
		defer response.Body.Close()

		// TODO: check the response status code and handle things appropriately
		// TODO: there is mention of different codes other than 200... how to handle these correctly?
		if response.StatusCode != http.StatusOK {
			log.Printf("🟠 warn: response had status code %d", response.StatusCode)
			continue
		}

		contentType := response.Header.Get("Content-Type")

		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			log.Printf("🟠 warn: failed to parse media type from content type: %v", err)
			// TODO: think about if i want to continue here
			continue
		}

		// fileExtensions, err := mime.ExtensionsByType(mediaType)
		// if err != nil {
		// 	log.Printf("warn: failed to get extensions from the mime type: %v", err)
		// 	continue
		// }
		// fmt.Println("DEBUG | fileExtensions", fileExtensions)

		// TODO: the above^ gives an array of possible extensions
		// but in this case fileExtensions[0] is ".jpe" which is weird
		// so I can't rely on the first index value being the file extension I want/expect

		var fileExtension string

		switch mediaType {
		case "image/jpeg":
			fileExtension = ".jpg"
			// TODO: extend this with other cases for other image file types
			// (although ideally it would be better to rely on the method shown above)
		default:
			// if we reach here it is not safe to proceed
			// because we will be copying the response body data into a file on the OS
			// which is dangerous if it is malicious
			// (also i can't break inside the switch/case)
			continue
		}

		parsedUrl, err := url.Parse(imageUrl)
		if err != nil {
			log.Printf("🟠 warn: cannot parse the image url: %v", err)
			// TODO: if we cannot parse the image url we should skip... right?
			continue
		}

		path := parsedUrl.Path
		compositeParts := strings.Split(path, "/")
		fileName := compositeParts[len(compositeParts)-1]

		outputFilepath := filepath.Join(temporaryDownloadsDirectory, fileName+fileExtension)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 3] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], outputFilepath)

		temporaryFile, err := os.Create(outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to create a temporary image file: %v", err)
			// TODO: think about if i want to continue here
			continue
		}
		defer temporaryFile.Close()

		_, err = io.Copy(temporaryFile, response.Body)
		if err != nil {
			log.Printf("🟠 warn: failed to save image %d\n with url %s\n", index+1, imageUrl)
			// TODO: think about if i want to continue here
			continue
		}

		imageDownloadCount++
	}

	log.Printf("🟢 success: downloaded %d images from the image urls", imageDownloadCount)

	// TODO: implement returning actual errors above in specific cases (but need to work out which)
	return nil
}
