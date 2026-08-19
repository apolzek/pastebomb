package shared

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Page is a clamped pagination window. Constructing it through NewPage makes
// an unbounded query impossible by construction.
type Page struct {
	number int
	size   int
}

// NewPage clamps the requested window into a range the database can serve
// cheaply: page numbers start at 1 and the size never exceeds maxPageSize.
func NewPage(number, size int) Page {
	if number < 1 {
		number = 1
	}
	switch {
	case size < 1:
		size = defaultPageSize
	case size > maxPageSize:
		size = maxPageSize
	}
	return Page{number: number, size: size}
}

func (p Page) Number() int { return p.number }
func (p Page) Size() int   { return p.size }
func (p Page) Limit() int  { return p.size }
func (p Page) Offset() int { return (p.number - 1) * p.size }
