package messageimpl

import "encoding/json"

type NewSignedCorpCLA struct {
	LinkId    string `json:"link_id"`
	SigningId string `json:"signing_id"`
	Email     string `json:"email"`
}

func (msg *NewSignedCorpCLA) message() ([]byte, error) {
	return json.Marshal(msg)
}

func UnmarshalToNewSignedCorpCLA(data []byte) (e NewSignedCorpCLA, err error) {
	err = json.Unmarshal(data, &e)

	return
}
