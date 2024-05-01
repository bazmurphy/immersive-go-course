package main

func GenerateDataMap(parsedImageUrlObjects []ParsedImageUrlObject, downloadedImageObjects []DownloadedImageObject, convertedImageObjects []ConvertedImageObject, uploadedImageObjects []UploadedImageObject) map[int][]string {
	// TODO: this map is built at the end of all the operations
	// it could be updated throughout as operations occur
	// but then I would be going back to the older method (that I just spent all this time refactoring out)

	dataMap := make(map[int][]string)

	for _, parsedImageUrlObject := range parsedImageUrlObjects {
		id := parsedImageUrlObject.ID
		dataMap[id] = append(dataMap[id], parsedImageUrlObject.ImageUrl)
	}

	for _, downloadedImageObject := range downloadedImageObjects {
		id := downloadedImageObject.ID
		dataMap[id] = append(dataMap[id], downloadedImageObject.ImageFilepath)
	}

	for _, convertedImageObject := range convertedImageObjects {
		id := convertedImageObject.ID
		dataMap[id] = append(dataMap[id], convertedImageObject.ImageFilepath)
	}

	for _, uploadedImageObject := range uploadedImageObjects {
		id := uploadedImageObject.ID
		dataMap[id] = append(dataMap[id], uploadedImageObject.ImageUrl)
	}

	return dataMap
}
