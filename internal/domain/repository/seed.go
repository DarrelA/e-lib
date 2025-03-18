package repository

type SeedRepository interface {
	SeedBooks() error
}
