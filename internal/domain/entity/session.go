package entity

type Session struct {
	UserID    string `json:"userID"`
	CreatedAt string `json:"createdAt"` // store CreatedAt as Unix timestamp string
}
