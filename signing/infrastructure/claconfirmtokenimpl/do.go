package claconfirmtokenimpl

import (
	"encoding/json"
)

type claConfirmTokenDO struct {
	LinkId   string `json:"link_id"`
	Email    string `json:"email"`
	NewCLAId string `json:"new_cla_id"`
	IssuedAt int64  `json:"issued_at"`
}

// MarshalBinary lets the struct be stored directly in redis.
func (do *claConfirmTokenDO) MarshalBinary() ([]byte, error) {
	return json.Marshal(do)
}

// UnmarshalBinary lets the struct be loaded directly from redis.
func (do *claConfirmTokenDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, do)
}

// marshalValue serializes a stored value the same way the redis client does:
// []byte/string as-is, other types as JSON.
func marshalValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case *claConfirmTokenDO:
		b, _ := v.MarshalBinary()
		return string(b)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
