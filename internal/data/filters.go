package data

import (
	"math"
	"strings"

	"greenlight.alexedwards.net/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// Define a new Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
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

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// metadata func calculates appropriate pagination metadata
// values given the total number of records, current page, page size
// last value is using math.ceil func to round float up to nth int
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		//we return a empty metadata struct if no records
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
