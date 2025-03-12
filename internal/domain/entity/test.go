package entity

import "time"

type Expected struct {
	Id              int    `json:"id"` // Primary key
	Method          string `json:"method"`
	UrlPath         string `json:"url_path"`
	StatusCode      int    `json:"status_code"`
	ResBodyContains string `json:"res_body_contains"`
}

type Actual struct {
	Id                int // Primary key
	ExpectedId        int // Foreign key
	StatusCode        int
	ReqUrlQueryString string
	ReqBody           string
	ResBody           string
	CreatedAt         time.Time
}
