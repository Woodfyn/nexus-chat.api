package image

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func SaveImage(image []byte, uploadDir string) (string, error) {
	fileName := uuid.New().String() + ".jpg"

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(image))
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func GetImage(imageName string, uploadDir string) ([]byte, error) {
	filePath := filepath.Join(uploadDir, imageName)
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return imageData, nil
}

func DeleteImage(imageName string, uploadDir string) error {
	filePath := filepath.Join(uploadDir, imageName)
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}
