package adapter

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

type fakeIndividualSigningService struct {
	app.IndividualSigningService

	confirmResult app.CLAConfirmResultDTO
	confirmErr    error
	gotToken      string
}

func (f *fakeIndividualSigningService) ConfirmByToken(cmd *app.CmdToConfirmCLAByToken) (app.CLAConfirmResultDTO, error) {
	f.gotToken = cmd.Token

	if f.confirmErr != nil {
		return app.CLAConfirmResultDTO{}, f.confirmErr
	}

	return f.confirmResult, nil
}

func TestAdapterConfirmByTokenSuccess(t *testing.T) {
	s := &fakeIndividualSigningService{
		confirmResult: app.CLAConfirmResultDTO{
			Result:      "confirmed",
			OrgAlias:    "openeuler",
			EmailMasked: "alic***@example.com",
			LinkId:      "link1",
			ClaId:       "9",
		},
	}

	v, merr := NewIndividualSigningAdapter(s).ConfirmByToken("tok")
	if merr != nil {
		t.Fatalf("unexpected model error: %v", merr)
	}
	if s.gotToken != "tok" {
		t.Fatalf("token passed to service = %q", s.gotToken)
	}

	if v.Result != "confirmed" || v.OrgAlias != "openeuler" || v.EmailMasked != "alic***@example.com" {
		t.Fatalf("unexpected result: %+v", v)
	}
	if v.LinkId != "link1" || v.ClaId != "9" {
		t.Fatalf("audit context lost: %+v", v)
	}
}

func TestAdapterConfirmByTokenErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want models.ModelErrCode
	}{
		{
			"invalid token",
			domain.NewDomainError(domain.ErrorCodeCLAConfirmTokenInvalid),
			models.ErrCLAConfirmTokenInvalid,
		},
		{
			"already latest",
			domain.NewDomainError(domain.ErrorCodeIndividualSigningCLAIsLatest),
			models.ErrCLAIsLatest,
		},
		{
			"system error",
			errors.New("db down"),
			models.ErrSystemError,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &fakeIndividualSigningService{confirmErr: c.err}

			_, merr := NewIndividualSigningAdapter(s).ConfirmByToken("tok")
			if merr == nil {
				t.Fatal("expected a model error")
			}
			if !merr.IsErrorOf(c.want) {
				t.Fatalf("error code = %s, want %s", merr.ErrCode(), c.want)
			}
		})
	}
}
