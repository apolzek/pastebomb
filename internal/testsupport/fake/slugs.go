package fake

import (
	"fmt"
	"sync/atomic"

	"github.com/apolzek/pastebomb/internal/domain/paste"
)

// SlugGenerator hands out predictable slugs so assertions can name them.
type SlugGenerator struct {
	prefix  string
	counter atomic.Int64
}

func NewSlugGenerator(prefix string) *SlugGenerator {
	return &SlugGenerator{prefix: prefix}
}

func (g *SlugGenerator) Generate() paste.Slug {
	return paste.Slug(fmt.Sprintf("%s-%d", g.prefix, g.counter.Add(1)))
}

// StaticSlugGenerator always returns the same slug, which is how a collision
// is simulated.
type StaticSlugGenerator struct{ Value string }

func (g StaticSlugGenerator) Generate() paste.Slug { return paste.Slug(g.Value) }
