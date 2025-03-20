package entity

import "time"

type Actual struct {
	ID                int // Primary key
	ExpectedID        int // Foreign key
	StatusCode        int
	ReqUrlQueryString string
	ReqBody           string
	ResBody           string
	CreatedAt         time.Time
}
