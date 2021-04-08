package mongodb

import (
	"encoding/hex"
	"encoding/json"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func newEncryption(key, nonce string) (encryption, error) {
	e := encryption{}

	se, err := util.NewSymmetricEncryption(key, nonce)
	if err != nil {
		return e, err
	}

	e.se = se
	return e, nil
}

type encryption struct {
	se util.SymmetricEncryption
}

func (e encryption) encryptBytes(data []byte) ([]byte, dbmodels.IDBError) {
	d, err := e.se.Encrypt(data)
	if err != nil {
		return nil, newSystemError(err)
	}
	return d, nil
}

func (e encryption) decryptBytes(data []byte) ([]byte, dbmodels.IDBError) {
	s, err := e.se.Decrypt(data)
	if err != nil {
		return nil, newSystemError(err)
	}
	return s, nil
}

func (e encryption) encryptStr(data string) (string, dbmodels.IDBError) {
	d, err := e.encryptBytes([]byte(data))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(d), nil
}

func (e encryption) decryptStr(data string) (string, dbmodels.IDBError) {
	b, err := hex.DecodeString(data)
	if err != nil {
		return "", newSystemError(err)
	}

	s, err1 := e.decryptBytes(b)
	if err1 != nil {
		return "", err1
	}

	return string(s), nil
}

func (e encryption) encryptSigningInfo(data *dbmodels.TypeSigningInfo) ([]byte, dbmodels.IDBError) {
	b, err := json.Marshal(*data)
	if err != nil {
		return nil, newSystemError(err)
	}

	return e.encryptBytes(b)
}

func (e encryption) decryptSigningInfo(data []byte) (*dbmodels.TypeSigningInfo, dbmodels.IDBError) {
	b, err := e.decryptBytes(data)
	if err != nil {
		return nil, err
	}

	var d dbmodels.TypeSigningInfo

	if err := json.Unmarshal(b, &d); err != nil {
		return nil, newSystemError(err)
	}
	return &d, nil
}
