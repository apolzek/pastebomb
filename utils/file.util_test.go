package utils

import (
	"os"
	"testing"
	"time"
)

func TestRemoveFile(t *testing.T) {
	// Crie um arquivo simulado para teste
	file, err := os.Create("../public/files/testfile.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	time.Sleep(2 * time.Second)
	// Chame a função RemoveFile para remover o arquivo
	err = RemoveFile("../public/files/testfile.txt")
	if err != nil {
		t.Errorf("RemoveFile failed, expected no error")
	}

	// Verifique se o arquivo foi removido corretamente
	if _, err := os.Stat("./public/files/testfile.txt"); !os.IsNotExist(err) {
		t.Errorf("File was not removed")
	}
}
