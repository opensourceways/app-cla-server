package randomstr

type RandomStr interface {
	New() (string, error)
}
