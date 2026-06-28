package utils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeContent(t *testing.T) {
	cases := []string{"", "hello", "pastebomb \xf0\x9f\x92\xa3", "linha 1\nlinha 2"}

	for _, in := range cases {
		got := EncodeContent(in)
		assert.NotNil(t, got)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(in)), *got)

		decoded, err := base64.StdEncoding.DecodeString(*got)
		assert.NoError(t, err)
		assert.Equal(t, in, string(decoded))
	}
}
