package Repository

import (
	"fmt"
	"math"
	"strings"
)

type Pagination struct {
	PagesItemsTotal int `json:"pagesItemsTotal"`
	PagesItemsLimit int `json:"pagesItemsLimit"`
	PagesTotal      int `json:"pagesTotal"`
	PageNumber      int `json:"pageNumber"`
}

func GetFilters(queryParams map[string][]string) (string, string) {

	filteredExact := make(map[string][]string)
	filteredRelated := make(map[string][]string)

	for key, value := range queryParams {
		if strings.HasPrefix(key, "filters[\"") || strings.HasPrefix(key, "filters[") {
			filterKey := key

			if strings.HasPrefix(key, "filters[\"@") {
				filterKey = strings.TrimPrefix(filterKey, "filters[\"@")
				filterKey = strings.TrimSuffix(filterKey, "\"]")
				filteredExact[filterKey] = value
			} else if strings.HasPrefix(key, "filters[@") {
				filterKey = strings.TrimPrefix(filterKey, "filters[@")
				filterKey = strings.TrimSuffix(filterKey, "]")
				filteredExact[filterKey] = value
			} else if strings.HasPrefix(key, "filters[\"") {
				filterKey = strings.TrimPrefix(filterKey, "filters[\"")
				filterKey = strings.TrimSuffix(filterKey, "\"]")
				filteredRelated[filterKey] = value
			} else {
				filterKey = strings.TrimPrefix(filterKey, "filters[")
				filterKey = strings.TrimSuffix(filterKey, "]")
				filteredRelated[filterKey] = value
			}
		}
	}

	exactStrings := make([]string, 0)
	relatedStrings := make([]string, 0)

	for key, value := range filteredExact {
		templateString := fmt.Sprintf("%s = '%s'", key, strings.Join(value, ", "))
		exactStrings = append(exactStrings, templateString)
	}

	for key, value := range filteredRelated {
		relatedString := fmt.Sprintf("CAST(%s AS VARCHAR) LIKE '%%%s%%'", key, strings.Join(value, ", "))
		relatedStrings = append(relatedStrings, relatedString)
	}

	filtersExact := strings.Join(exactStrings, " AND ")
	filtersRelated := strings.Join(relatedStrings, " OR ")

	return filtersExact, filtersRelated
}

func GetPagination(page int, limit int, totalItems int64) Pagination {
	pageTotal := int(math.Ceil(float64(totalItems) / float64(limit)))

	pagination := Pagination{
		PagesItemsTotal: int(totalItems),
		PagesItemsLimit: limit,
		PagesTotal:      pageTotal,
		PageNumber:      page,
	}
	return pagination
}
