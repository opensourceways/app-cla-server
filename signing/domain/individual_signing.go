package domain

import "github.com/opensourceways/app-cla-server/util"

const (
	individualSigningActionSign  = "sign"
	individualSigningActionAgree = "agree"
)

type IndividualSigning struct {
	Id      string
	Link    LinkInfo
	Rep     Representative
	Logs    []IndividualSigningLog
	Date    string
	AllInfo AllSingingInfo
	Version int
}

type IndividualSigningLog struct {
	Date   string
	ClaId  string
	Action string
}

func NewIndividualSigning(link LinkInfo, rep Representative, all AllSingingInfo) IndividualSigning {
	is := IndividualSigning{
		Link:    link,
		Rep:     rep,
		Date:    util.Date(),
		AllInfo: all,
	}

	is.addLogOfSigning()

	return is
}

func (i *IndividualSigning) AgreeNewCLA(claId string) {
	i.Link.CLAId = claId
	i.addLogOfAgreeingNewCLA(util.Date(), claId)
}

func (i *IndividualSigning) addLogOfSigning() {
	i.Logs = []IndividualSigningLog{
		{
			Date:   i.Date,
			ClaId:  i.Link.CLAId,
			Action: individualSigningActionSign,
		},
	}
}

func (i *IndividualSigning) addLogOfAgreeingNewCLA(date, claId string) {
	if len(i.Logs) == 0 {
		i.addLogOfSigning()
	}

	i.Logs = append(i.Logs, IndividualSigningLog{
		Date:   date,
		ClaId:  claId,
		Action: individualSigningActionAgree,
	})
}
