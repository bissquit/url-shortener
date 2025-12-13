package service

type IDGenerator interface {
	GenerateShortID() (string, error)
}
