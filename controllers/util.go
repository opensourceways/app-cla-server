package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
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

func notifyCorpAdmin(linkId string, orgInfo *models.OrgInfo, info *models.CorporationManagerCreateOption) {
	notifyCorpManagerWhenAdding(linkId, orgInfo, []models.CorporationManagerCreateOption{*info})
}

func notifyCorpManagerWhenAdding(linkId string, orgInfo *models.OrgInfo, info []models.CorporationManagerCreateOption) {
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
			ProjectURL:       orgInfo.ProjectURL,
			URLOfCLAPlatform: config.signingURL(linkId),
		}

		sendEmailToIndividual(item.Email, orgInfo, subject, &d)

		// clear password
		for i := range item.Password {
			item.Password[i] = 0
		}
	}
}

func notifyCorpManagerWhenCLAUpdated(userId, linkId string) {
	summary, err := models.ListCorpSigning(userId, linkId)
	if err != nil {
		logs.Error("list corp signing failed when cla updated:", err)

		return
	}

	org, err := models.GetLink(linkId)
	if err != nil {
		logs.Error("get org info failed when cla updated:", err)

		return
	}

	for _, v := range summary {
		b := emailtmpl.CLAUpdate{
			Org:              org.OrgAlias,
			AdminName:        v.AdminName,
			ProjectURL:       org.ProjectURL,
			URLOfCLAPlatform: config.signingURL(linkId),
		}

		sendEmail([]string{v.AdminEmail}, &org, "CLA has been updated", &b)
	}
}

func fetchInputPayloadData(input []byte, info interface{}) *failedApiResult {
	if err := json.Unmarshal(input, info); err != nil {
		return newFailedApiResult(
			400, errParsingApiBody, fmt.Errorf("invalid input payload: %s", err.Error()),
		)
	}
	return nil
}

func genCLADiff(diffFile string) {
	_, err := os.Stat(diffFile)
	if err == nil {
		// diff pdf is exists
		return
	}

	fileName := filepath.Base(diffFile)
	split := strings.Split(strings.TrimSuffix(fileName, ".pdf"), "_")
	if len(split) != 3 {
		logs.Error("generate diff pdf failed, file path is invalid")

		return
	}

	linkId := split[0]
	oldClaId := split[1]
	newClaId := split[2]

	oldPDFPath := models.CLAFile(linkId, oldClaId)
	newPDFPath := models.CLAFile(linkId, newClaId)

	err = pdf.GetPDFGenerator().GenPDFDiff(oldPDFPath, newPDFPath, diffFile)
	if err != nil {
		logs.Error("generate diff pdf failed:", diffFile, err)
	}
}
