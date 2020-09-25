package mongodb

import (
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func structToMap(info interface{}) (map[string]interface{}, error) {
	body, err := golangsdk.BuildRequestBody(info, "")
	if err != nil {
		return nil, dbmodels.DBError{
			ErrCode: dbmodels.ErrInvalidParameter,
			Err:     err,
		}
	}
	return body, nil
}

func addCorporationID(email string, body map[string]interface{}) {
	body[fieldCorporationID] = util.EmailSuffix(email)
}
