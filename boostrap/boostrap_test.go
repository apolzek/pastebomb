package boostrap

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/tj/assert"
)

func TestBootstrapAppLoadEnv(t *testing.T) {

	err := godotenv.Load("../.env")
	assert.NoError(t, err, "Error loading environment variables")

}
