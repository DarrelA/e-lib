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
	TestId            int
	ExpectedId        int // Foreign key
	StatusCode        int
	ReqUrlQueryString string
	ReqBody           string
	ResBody           string
	CreatedAt         time.Time
}

type CompareResult struct {
	Id                 int       `json:"id"`
	Method             string    `json:"method"`
	UrlPath            string    `json:"url_path"`
	ReqUrlQueryString  string    `json:"req_url_query_string"`
	ReqBody            string    `json:"req_body"`
	ExpectedStatusCode int       `json:"expected_status_code"`
	ActualStatusCode   int       `json:"actual_status_code"`
	ResBody            string    `json:"res_body"`
	ResBodyContains    string    `json:"res_body_contains"`
	CreatedAt          time.Time `json:"created_at"`
	Reason             []string  `json:"reason"`
	Pass               bool      `json:"pass"`
}
