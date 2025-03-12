package entity

import "time"

type Actual struct {
	Id                int // Primary key
	ExpectedId        int // Foreign key
	StatusCode        int
	ReqUrlQueryString string
	ReqBody           string
	ResBody           string
	CreatedAt         time.Time
}
