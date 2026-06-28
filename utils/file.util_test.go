package utils

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "testfile.txt")
	f, err := os.Create(path)
	assert.NoError(t, err)
	f.Close()

	assert.NoError(t, RemoveFile(path))

	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "file should have been removed")
}

func TestRemoveFile_NotFound(t *testing.T) {
	err := RemoveFile(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	assert.Error(t, err)
}

func TestRandomString(t *testing.T) {
	const expected = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s := RandomString(12)
	assert.Len(t, s, 12)
	for _, c := range s {
		assert.True(t, strings.ContainsRune(expected, c))
	}
}

func TestRandomFileName_DefaultPrefix(t *testing.T) {
	name := RandomFileName(".jpg")
	assert.True(t, strings.HasPrefix(name, "file-"))
	assert.True(t, strings.HasSuffix(name, ".jpg"))
}

func TestRandomFileName_CustomPrefix(t *testing.T) {
	name := RandomFileName(".png", "avatar")
	assert.True(t, strings.HasPrefix(name, "avatar-"))
	assert.True(t, strings.HasSuffix(name, ".png"))
}

func TestRandomFileName_EmptyPrefixFallsBack(t *testing.T) {
	name := RandomFileName(".txt", "")
	assert.True(t, strings.HasPrefix(name, "file-"))
}

func TestFileValidationByExtension(t *testing.T) {
	allowed := []string{".jpg", ".png"}

	assert.True(t, FileValidationByExtension(&multipart.FileHeader{Filename: "photo.jpg"}, allowed))
	assert.True(t, FileValidationByExtension(&multipart.FileHeader{Filename: "pic.png"}, allowed))
	assert.False(t, FileValidationByExtension(&multipart.FileHeader{Filename: "doc.pdf"}, allowed))
	assert.False(t, FileValidationByExtension(&multipart.FileHeader{Filename: "noext"}, allowed))
}
