package image_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Woodfyn/chat-api-backend-go/pkg/image"
)

func TestUtils_saveImage(t *testing.T) {
	tests := []struct {
		name      string
		image     []byte
		uploadDir string
		wantErr   bool
	}{
		{
			name:      "ok",
			image:     []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x64, 0x08, 0x06, 0x00, 0x00, 0x00},
			uploadDir: "./static-test",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := image.SaveImage(tt.image, tt.uploadDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveImage() error = %v, wantErr %v", err, tt.wantErr)
			}

			defer func() {
				filePath := filepath.Join(tt.uploadDir, "*.jpg")
				matches, err := filepath.Glob(filePath)
				if err != nil {
					t.Errorf("Error when searching for the file: %v", err)
				}

				for _, match := range matches {
					err := os.Remove(match)
					if err != nil {
						t.Errorf("Error deleting file: %v", err)
					}
					err = os.Remove(tt.uploadDir)
					if err != nil {
						t.Errorf("Error deleting directory: %v", err)
					}
				}
			}()

			filePath := filepath.Join(tt.uploadDir, "*.jpg")
			matches, err := filepath.Glob(filePath)
			if err != nil {
				t.Errorf("Error when searching for the file: %v", err)
			}

			if len(matches) != 1 {
				t.Errorf("Expected 1 file, got %d", len(matches))
			}

			savedData, err := os.ReadFile(matches[0])
			if err != nil {
				t.Errorf("Error when reading the saved file: %v", err)
			}

			if !bytes.Equal(savedData, tt.image) {
				t.Error("Saved image data does not match original image data")
			}
		})
	}
}
