package randomcode

type RandomCode interface {
	New() (string, error)
	IsValid(string) bool
}
