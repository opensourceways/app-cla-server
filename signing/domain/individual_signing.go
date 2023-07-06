package domain

type IndividualSigning struct {
	Id      string
	Link    LinkInfo
	Rep     Representative
	Date    string
	AllInfo AllSingingInfo
}
