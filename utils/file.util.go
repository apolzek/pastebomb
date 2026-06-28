package utils

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// RandomString delegates to the crypto/rand-backed generator so file names
// and slugs cannot be predicted by callers.
func RandomString(n int) string {
	return GenerateRandomString(n)
}

func FileValidationByExtension(fileHeader *multipart.FileHeader, fileExtension []string) bool {
	extension := filepath.Ext(fileHeader.Filename)
	for _, typeFile := range fileExtension {
		if extension == typeFile {
			return true
		}
	}
	return false
}

func RandomFileName(extensionFile string, prefix ...string) string {
	currentPrefix := "file"
	if len(prefix) > 0 && prefix[0] != "" {
		currentPrefix = prefix[0]
	}
	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s-%s-%s%s", currentPrefix, currentTime, RandomString(5), extensionFile)
}

func SaveFile(ctx *gin.Context, fileHeader *multipart.FileHeader, fileName string) bool {
	dst := fmt.Sprintf("./public/files/%s", fileName)
	if err := ctx.SaveUploadedFile(fileHeader, dst); err != nil {
		slog.Error("upload failed", "dst", dst, "err", err)
		return false
	}
	return true
}

func RemoveFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		slog.Error("remove file failed", "path", filePath, "err", err)
		return err
	}
	return nil
}
