package postgres

type SeedRepository interface {
	SeedBooks(pathToBooksJsonFile string) error
}
