package security

import (
	"crypto/rand"
	"math/big"

	"github.com/apolzek/pastebomb/internal/domain/paste"
)

// slugAlphabet is base62. Ambiguity between similar glyphs is acceptable here
// because handles are copied, never transcribed by hand.
const slugAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomSlugGenerator draws handles from crypto/rand. Predictable handles
// would defeat the "unlisted paste" guarantee entirely, so math/rand is not an
// acceptable substitute here.
type RandomSlugGenerator struct{ length int }

func NewRandomSlugGenerator(length int) *RandomSlugGenerator {
	if length <= 0 {
		length = paste.SlugLength
	}
	return &RandomSlugGenerator{length: length}
}

// Generate returns a fresh handle. A failing system RNG is unrecoverable: any
// fallback would silently hand out guessable handles, so it panics instead.
func (g *RandomSlugGenerator) Generate() paste.Slug {
	upperBound := big.NewInt(int64(len(slugAlphabet)))
	out := make([]byte, g.length)
	for i := range out {
		index, err := rand.Int(rand.Reader, upperBound)
		if err != nil {
			panic("crypto/rand unavailable, refusing to generate a guessable slug: " + err.Error())
		}
		out[i] = slugAlphabet[index.Int64()]
	}
	return paste.Slug(out)
}
