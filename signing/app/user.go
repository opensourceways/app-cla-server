package app

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/captchaservice"
	"github.com/opensourceways/app-cla-server/signing/domain/loginservice"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/symmetricencryption"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewUserService(
	us userservice.UserService,
	ls loginservice.LoginService,
	repo repository.CorpSigning,
	encrypt symmetricencryption.Encryption,
	userRepo repository.User,
	interval time.Duration,
	vcService vcservice.VCService,
	captchaSvc captchaservice.CaptchaService,
	privacyVersion string,
) UserService {
	return &userService{
		us:             us,
		ls:             ls,
		repo:           repo,
		encrypt:        encrypt,
		userRepo:       userRepo,
		interval:       interval,
		vcService:      verificationCodeService{vcService},
		captchaService: captchaSvc,
		privacyVersion: privacyVersion,
	}
}

type UserService interface {
	Get(userId string) (dto UserBasicInfoDTO, err error)
	Login(cmd *CmdToLogin) (dto UserLoginDTO, err error)
	GetCaptcha() (id string, imageBase64 string, err error)
	ResetPassword(cmd *CmdToResetPassword) error
	ChangePassword(cmd *CmdToChangePassword) error
	GenKeyForPasswordRetrieval(*CmdToGenKeyForPasswordRetrieval) (string, error)
}

type userService struct {
	us             userservice.UserService
	ls             loginservice.LoginService
	repo           repository.CorpSigning
	encrypt        symmetricencryption.Encryption
	userRepo       repository.User
	interval       time.Duration
	vcService      verificationCodeService
	captchaService captchaservice.CaptchaService
	privacyVersion string
}

func (s *userService) ChangePassword(cmd *CmdToChangePassword) error {
	err := s.us.ChangePassword(cmd.Id, cmd.OldOne, cmd.NewOne)
	cmd.clear()

	return err
}

func (s *userService) GenKeyForPasswordRetrieval(cmd *CmdToGenKeyForPasswordRetrieval) (string, error) {
	b, err := s.us.IsAValidUser(cmd.Id, cmd.EmailAddr)
	if err != nil {
		return "", err
	}
	if !b {
		return "", domain.NewDomainError(domain.ErrorCodeUserNotExists)
	}

	code, err := s.vcService.newCodeIfItCan(cmd, s.interval)
	if err != nil {
		return "", err
	}

	k := resettingPasswordKey{
		Email: cmd.EmailAddr.EmailAddr(),
		Code:  code,
	}

	v, err := json.Marshal(k)
	if err != nil {
		return "", err
	}

	v, err = s.encrypt.Encrypt(v)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(v), nil
}

func (s *userService) ResetPassword(cmd *CmdToResetPassword) error {
	defer cmd.clear()

	v, err := hex.DecodeString(cmd.Key)
	if err != nil {
		return err
	}

	v, err = s.encrypt.Decrypt(v)
	if err != nil {
		return err
	}

	k := resettingPasswordKey{}

	if err := json.Unmarshal(v, &k); err != nil {
		return err
	}

	e, err := k.toEmail()
	if err != nil {
		return err
	}

	err = s.vcService.validate(
		&CmdToGenKeyForPasswordRetrieval{
			Id:        cmd.LinkId,
			EmailAddr: e,
		},
		k.Code,
	)
	if err != nil {
		return err
	}

	return s.us.ResetPassword(cmd.LinkId, e, cmd.NewOne)
}

func (s *userService) GetCaptcha() (id string, imageBase64 string, err error) {
	return s.captchaService.Generate()
}

func (s *userService) Login(cmd *CmdToLogin) (dto UserLoginDTO, err error) {
	defer cmd.clear()

	// Determine the login identifier (account name or email address).
	var lid string
	if cmd.Account != nil {
		lid = cmd.Account.Account()
	} else {
		lid = cmd.Email.EmailAddr()
	}

	// Check whether captcha verification is required before attempting login.
	if needed, checkErr := s.ls.NeedCaptcha(lid); checkErr != nil {
		logs.Warn("failed to check captcha requirement, err: %s", checkErr.Error())
	} else if needed {
		if verifyErr := s.captchaService.Verify(cmd.CaptchaId, cmd.CaptchaAnswer); verifyErr != nil {
			dto.NeedCaptcha = true
			err = domain.NewDomainError(domain.ErrorCodeCaptchaInvalid)
			return
		}
	}

	var u domain.User
	var l domain.Login

	if cmd.Account != nil {
		u, l, err = s.ls.LoginByAccount(cmd.LinkId, cmd.Account, cmd.Password)
	} else {
		u, l, err = s.ls.LoginByEmail(cmd.LinkId, cmd.Email, cmd.Password)
	}

	// It should record the retry number whatever if it is success or not.
	dto.RetryNum = l.RetryNum()
	dto.NeedCaptcha = l.NeedCaptcha()

	if err != nil {
		return
	}

	if err = s.checkPrivacyConsent(cmd.PrivacyConsented, &u); err != nil {
		return
	}

	if dto.Role, err = s.getRole(&u); err != nil {
		return
	}

	dto.Email = u.EmailAddr.EmailAddr()
	dto.UserId = u.Id
	dto.CorpSigningId = u.CorpSigningId
	dto.PrivacyVersion = u.PrivacyConsent.Version
	dto.InitialPWChanged = u.PasswordChanged

	return
}

func (s *userService) checkPrivacyConsent(privacyConsented bool, u *domain.User) error {
	if !u.UpdatePrivacyConsent(s.privacyVersion) {
		return nil
	}

	if !privacyConsented {
		return domain.NewDomainError(domain.ErrorPrivacyConsentInvalid)
	}

	return s.userRepo.SavePrivacyConsent(u)
}

func (s *userService) Get(userId string) (dto UserBasicInfoDTO, err error) {
	u, err := s.us.Get(userId)
	if err != nil {
		return
	}

	dto.UserId = u.Account.Account()
	dto.InitialPWChanged = u.PasswordChanged

	dto.Role, err = s.getRole(&u)

	return
}

func (s *userService) getRole(u *domain.User) (string, error) {
	if u.IsCommunityManager() {
		return "", nil
	}

	cs, err := s.repo.Find(u.CorpSigningId)
	if err != nil {
		return "", err
	}

	if role := cs.GetRole(u.EmailAddr); role != "" {
		return role, nil
	}

	return "", errors.New("no role")
}
