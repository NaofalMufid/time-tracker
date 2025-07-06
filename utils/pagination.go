package utils

import (
	"math"
	"net/http"
	"strconv"
)

type PageDetails struct {
	TotalRecords int64 `json:"totalRecords"`
	TotalPages   int   `json:"totalPages"`
	CurrentPage  int   `json:"currentPage"`
	PageSize     int   `json:"pageSize"`
	HasNext      bool  `json:"hasNext"`
	HasPrev      bool  `json:"hasPrev"`
}

func GetParams(r *http.Request) (page, pageSize int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		pageSize = 10
	}

	return page, pageSize
}

func CalculateMetadata(totalRecords int64, page, pageSize int) PageDetails {
	if totalRecords == 0 {
		return PageDetails{}
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	return PageDetails{
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  page,
		PageSize:     pageSize,
		HasNext:      page < totalPages,
		HasPrev:      page > 1,
	}
}
