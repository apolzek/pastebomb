package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/infrastructure/security"
)

func TestGenerateProducesSlugsOfTheRequestedLength(t *testing.T) {
	t.Parallel()

	generator := security.NewRandomSlugGenerator(16)
	assert.Len(t, generator.Generate().String(), 16)
}

func TestNonPositiveLengthFallsBackToTheDomainDefault(t *testing.T) {
	t.Parallel()

	assert.Len(t, security.NewRandomSlugGenerator(0).Generate().String(), paste.SlugLength)
	assert.Len(t, security.NewRandomSlugGenerator(-5).Generate().String(), paste.SlugLength)
}

// Slugs are the only thing protecting an unlisted paste, so a repeat inside a
// small sample would be disqualifying.
func TestGeneratedSlugsDoNotRepeat(t *testing.T) {
	t.Parallel()

	generator := security.NewRandomSlugGenerator(paste.SlugLength)
	seen := make(map[string]struct{}, 10_000)

	for range 10_000 {
		slug := generator.Generate().String()
		_, duplicate := seen[slug]
		assert.False(t, duplicate, "slug %q was generated twice", slug)
		seen[slug] = struct{}{}
	}
}

func TestGeneratedSlugsUseOnlyURLSafeCharacters(t *testing.T) {
	t.Parallel()

	generator := security.NewRandomSlugGenerator(paste.SlugLength)

	for range 200 {
		for _, char := range generator.Generate().String() {
			isAlphanumeric := (char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9')
			assert.True(t, isAlphanumeric, "unexpected character %q in slug", char)
		}
	}
}
