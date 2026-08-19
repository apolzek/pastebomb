package memory

import "github.com/apolzek/pastebomb/internal/domain/shared"

// paginate applies the same window the SQL adapters apply, so a test that
// passes against these fakes exercises the same slicing behaviour.
func paginate[T any](items []T, page shared.Page) []T {
	offset := page.Offset()
	if offset >= len(items) {
		return nil
	}
	end := min(offset+page.Limit(), len(items))
	return items[offset:end]
}
