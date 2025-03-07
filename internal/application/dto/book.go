package dto

type BookTitleAvailability struct {
	Title           string `json:"title"`
	AvailableCopies int    `json:"available_copies"`
}

type BorrowBook struct {
	Title string `json:"title"`
}
