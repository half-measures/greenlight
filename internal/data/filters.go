package data

import (
	"strings"

	"greenlight.alexedwards.net/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	//check page and pagesize param are good
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a max of 10million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a max of 100")

	// check sort param matches value in our safelist
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// This func checks the client provided sort field matches a entry
// if it does, extract column name from sot field with stripping
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)

}

// below is how we sort our API results
// return sort direction ASC / DESC depending on prefix char of sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
