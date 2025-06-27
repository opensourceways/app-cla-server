package message

type CLAUpdatedMsg struct {
	LinkId   string
	OldCLAId string
	NewCLAId string
}

type Message interface {
	CLAUpdated(CLAUpdatedMsg)
}
