package randombytes

type RandomBytes interface {
	New(int) ([]byte, error)
}
