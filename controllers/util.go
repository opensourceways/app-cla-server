package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/worker"
)

const (
	headerToken                    = "Token"
	headerPasswordRetrievalKey     = "Password-Retrieval-Key"
	apiAccessController            = "access_controller"
	fileNameOfUploadingOrgSignatue = "org_signature_file"
)

func sendEmailToIndividual(to string, orgInfo *models.OrgInfo, subject string, builder emailservice.IEmailMessageBulder) {
	sendEmail([]string{to}, orgInfo, subject, builder)
}

func sendEmail(to []string, orgInfo *models.OrgInfo, subject string, builder emailservice.IEmailMessageBulder) {
	msg, err := builder.GenEmailMsg()
	if err != nil {
		logs.Error(err)
		return
	}
	msg.From = orgInfo.OrgEmail
	msg.To = to
	msg.Subject = subject

	worker.GetEmailWorker().SendSimpleMessage(orgInfo.OrgEmailPlatform, &msg)
}

func notifyCorpAdmin(orgInfo *models.OrgInfo, info *models.CorporationManagerCreateOption) {
	notifyCorpManagerWhenAdding(orgInfo, []models.CorporationManagerCreateOption{*info})
}

func notifyCorpManagerWhenAdding(orgInfo *models.OrgInfo, info []models.CorporationManagerCreateOption) {
	admin := info[0].Role == models.RoleAdmin
	subject := fmt.Sprintf("Account on project of \"%s\"", orgInfo.OrgAlias)

	for i := range info {
		item := &info[i]
		d := emailtmpl.AddingCorpManager{
			Admin:            admin,
			ID:               item.ID,
			User:             item.Name,
			Email:            item.Email,
			Password:         item.Password,
			Org:              orgInfo.OrgAlias,
			ProjectURL:       orgInfo.ProjectURL(),
			URLOfCLAPlatform: config.CLAPlatformURL,
		}

		sendEmailToIndividual(item.Email, orgInfo, subject, &d)

		// clear password
		for i := range item.Password {
			item.Password[i] = 0
		}
	}
}

func getSingingInfo(info models.TypeSigningInfo, fields []models.CLAField) models.TypeSigningInfo {
	if len(info) == 0 {
		return info
	}

	r := models.TypeSigningInfo{}
	for i := range fields {
		fid := fields[i].ID
		if v, ok := info[fid]; ok {
			r[fid] = v
		}
	}
	return r
}

func fetchInputPayloadData(input []byte, info interface{}) *failedApiResult {
	if err := json.Unmarshal(input, info); err != nil {
		return newFailedApiResult(
			400, errParsingApiBody, fmt.Errorf("invalid input payload: %s", err.Error()),
		)
	}
	return nil
}
