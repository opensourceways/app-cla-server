package domain

type IndividualSigning struct {
	Id      string
	Link    Link
	Rep     Representative
	Date    string
	AllInfo AllSingingInfo
}
