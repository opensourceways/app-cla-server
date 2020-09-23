package mongodb

import (
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/dbmodels"
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
