package postgres

type SeedRepository interface {
	SeedBooks() error
	CompareTestReqAndRes()
}
