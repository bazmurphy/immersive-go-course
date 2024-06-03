package api

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Image struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
}

func GetAllImages(database *pgx.Conn) ([]Image, error) {
	query := "SELECT title, url, alt_text FROM images"

	rows, err := database.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("error: failed to query the database for all images: %w", err)
	}

	defer rows.Close()

	var images []Image

	for rows.Next() {
		var image Image

		err := rows.Scan(&image.Title, &image.URL, &image.AltText)
		if err != nil {
			return nil, fmt.Errorf("error: failed to read a row from the database: %w", err)
		}

		images = append(images, image)
	}

	return images, nil
}

func AddImage(database *pgx.Conn, newImage Image) (Image, error) {
	query := "INSERT INTO images (title, url, alt_text) VALUES ($1, $2, $3) RETURNING title, url, alt_text"

	var insertImage Image

	err := database.QueryRow(context.Background(), query, newImage.Title, newImage.URL, newImage.AltText).Scan(
		&insertImage.Title,
		&insertImage.URL,
		&insertImage.AltText,
	)
	if err != nil {
		return Image{}, fmt.Errorf("error: inserting new image into the database: %w", err)
	}

	return insertImage, nil
}
