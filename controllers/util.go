package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/util"
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

	worker.GetEmailWorker().SendSimpleMessage(orgInfo.OrgEmailPlatform, msg)
}

func notifyCorpAdmin(orgInfo *models.OrgInfo, info *dbmodels.CorporationManagerCreateOption) {
	notifyCorpManagerWhenAdding(orgInfo, []dbmodels.CorporationManagerCreateOption{*info})
}

func notifyCorpManagerWhenAdding(orgInfo *models.OrgInfo, info []dbmodels.CorporationManagerCreateOption) {
	admin := info[0].Role == dbmodels.RoleAdmin
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
			URLOfCLAPlatform: config.AppConfig.CLAPlatformURL,
		}

		sendEmailToIndividual(item.Email, orgInfo, subject, d)
	}
}

func getSingingInfo(info dbmodels.TypeSigningInfo, fields []dbmodels.Field) dbmodels.TypeSigningInfo {
	if len(info) == 0 {
		return info
	}

	r := dbmodels.TypeSigningInfo{}
	for i := range fields {
		fid := fields[i].ID
		if v, ok := info[fid]; ok {
			r[fid] = v
		}
	}
	return r
}

func parseOrgAndRepo(s string) (string, string) {
	v := strings.Split(s, ":")
	if len(v) == 2 {
		return v[0], v[1]
	}
	return s, ""
}

func buildOrgRepo(platform, orgID, repoID string) *models.OrgRepo {
	return &models.OrgRepo{
		Platform: platform,
		OrgID:    orgID,
		RepoID:   repoID,
	}
}

func genOrgFileLockPath(platform, org, repo string) string {
	return util.GenFilePath(
		config.AppConfig.PDFOrgSignatureDir,
		util.GenFileName("lock", platform, org, repo),
	)
}

func genCLAFilePath(linkID, applyTo, language, hash string) string {
	return util.GenFilePath(
		config.AppConfig.PDFOrgSignatureDir,
		util.GenFileName("cla", linkID, applyTo, language, hash, ".pdf"))
}

func genLinkID(v *dbmodels.OrgRepo) string {
	repo := ""
	if v.RepoID != "" {
		repo = fmt.Sprintf("_%s", v.RepoID)
	}
	return fmt.Sprintf("%s_%s%s-%d", v.Platform, v.OrgID, repo, time.Now().UnixNano())
}

func getCLAInfoSigned(linkID, claLang, applyTo string) (*models.CLAInfo, *failedApiResult) {
	claInfo, merr := models.GetCLAInfoSigned(linkID, claLang, applyTo)
	if merr == nil {
		if claInfo == nil {
			return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("cla info is empty, impossible"))
		}
		return claInfo, nil
	}

	if merr.IsErrorOf(models.ErrNoLinkOrUnsigned) {
		return nil, nil
	}
	return nil, parseModelError(merr)
}

func fetchInputPayloadData(input []byte, info interface{}) *failedApiResult {
	if err := json.Unmarshal(input, info); err != nil {
		return newFailedApiResult(
			400, errParsingApiBody, fmt.Errorf("invalid input payload: %s", err.Error()),
		)
	}
	return nil
}

func lockOnRepo(orgInfo *dbmodels.OrgInfo) (func(), *failedApiResult) {
	unlock, err := util.Lock(genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID))
	if err != nil {
		return nil, newFailedApiResult(500, errSystemError, err)
	}
	return unlock, nil
}

func listCorpEmailDomain(linkID, adminEmail string) (map[string]bool, *failedApiResult) {
	d, err := models.ListCorpEmailDomain(linkID, adminEmail)
	if err != nil {
		return nil, parseModelError(err)
	}
	if len(d) == 0 {
		return nil, newFailedApiResult(400, errUnsigned, fmt.Errorf("unsigned"))
	}

	m := map[string]bool{}
	for _, i := range d {
		m[i] = true
	}
	return m, nil
}
