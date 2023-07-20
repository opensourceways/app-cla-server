package app

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	vcTypeCorp              = "corp"
	vcTypeEmployee          = "employee"
	vcTypeIndividual        = "individual"
	vcTypeSMTPEmail         = "smtp email"
	vcTypeEmailDomain       = "email domain"
	vcTypePasswordRetrieval = "password retrieval"
)

type vcPurpose interface {
	purpose() (dp.Purpose, error)
}

// signing
type CmdToCreateVerificationCode struct {
	Id string // link id or corp signing id

	EmailAddr dp.EmailAddr
}

func (cmd *CmdToCreateVerificationCode) genPurpose(codeType string) (dp.Purpose, error) {
	return dp.NewPurpose(
		fmt.Sprintf("%s, %s, %s", cmd.Id, codeType, strings.ToLower(cmd.EmailAddr.EmailAddr())),
	)
}

// cmdToCreateCodeForCorpSigning
type cmdToCreateCodeForCorpSigning CmdToCreateVerificationCode

func (cmd *cmdToCreateCodeForCorpSigning) purpose() (dp.Purpose, error) {
	return (*CmdToCreateVerificationCode)(cmd).genPurpose(vcTypeCorp)
}

// cmdToCreateCodeForEmployeeSigning
type cmdToCreateCodeForEmployeeSigning CmdToCreateVerificationCode

func (cmd *cmdToCreateCodeForEmployeeSigning) purpose() (dp.Purpose, error) {
	return (*CmdToCreateVerificationCode)(cmd).genPurpose(vcTypeEmployee)
}

// cmdToCreateCodeForIndividualSigning
type cmdToCreateCodeForIndividualSigning CmdToCreateVerificationCode

func (cmd *cmdToCreateCodeForIndividualSigning) purpose() (dp.Purpose, error) {
	return (*CmdToCreateVerificationCode)(cmd).genPurpose(vcTypeIndividual)
}
