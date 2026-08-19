package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

func TestNewPageClampsOutOfRangeInput(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		number, size          int
		wantNumber, wantSize  int
		wantOffset, wantLimit int
	}{
		"defaults for zero values":   {0, 0, 1, 20, 0, 20},
		"negative page becomes one":  {-5, 10, 1, 10, 0, 10},
		"oversized page is capped":   {2, 5000, 2, 100, 100, 100},
		"offset follows page number": {3, 10, 3, 10, 20, 10},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			page := shared.NewPage(tc.number, tc.size)
			assert.Equal(t, tc.wantNumber, page.Number())
			assert.Equal(t, tc.wantSize, page.Size())
			assert.Equal(t, tc.wantOffset, page.Offset())
			assert.Equal(t, tc.wantLimit, page.Limit())
		})
	}
}
