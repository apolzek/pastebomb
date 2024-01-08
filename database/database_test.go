package database

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestConnectDatabase(t *testing.T) {
	err := godotenv.Load("../.env")
	assert.NoError(t, err, "Error loading environment variables")

	// Chame a função de inicialização
	ConnectDatabase()

	// Verifique se a conexão com o banco de dados foi bem-sucedida
	assert.NotNil(t, DB, "DB must not be null if the connection was successful")

}
