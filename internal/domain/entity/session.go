package entity

import "time"

type Session struct {
	UserID    int64
	CreatedAt time.Time
}
