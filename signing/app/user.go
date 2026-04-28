package app

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
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

	lid := s.getLoginId(cmd)

	if err := s.validateCaptcha(cmd, lid, &dto); err != nil {
		return dto, err
	}

	u, l, err := s.doLogin(cmd, lid)
	if err != nil {
		dto.RetryNum = l.RetryNum()
		dto.NeedCaptcha = l.NeedCaptcha()
		return dto, err
	}

	s.clearLoginFailure(lid)
	err = s.fillLoginResult(cmd, &dto, u)
	return dto, err
}

func (s *userService) getLoginId(cmd *CmdToLogin) string {
	if cmd.Account != nil {
		return cmd.Account.Account()
	}
	return cmd.Email.EmailAddr()
}

func (s *userService) validateCaptcha(cmd *CmdToLogin, lid string, dto *UserLoginDTO) error {
	loginInfo, loginInfoErr := s.ls.GetLoginInfo(lid)
	if loginInfoErr != nil {
		if commonRepo.IsErrorResourceNotFound(loginInfoErr) {
			return nil
		}
		logs.Error("failed to get login info for %s, err: %s", lid, loginInfoErr.Error())
		return loginInfoErr
	}

	needCaptcha := false
	if loginInfo != nil {
		dto.RetryNum = loginInfo.RetryNum()
		needCaptcha = loginInfo.NeedCaptcha()
		dto.NeedCaptcha = needCaptcha
	}

	hasCaptcha := cmd.CaptchaId != "" && cmd.CaptchaAnswer != ""

	if needCaptcha && !hasCaptcha {
		return domain.NewDomainError(domain.ErrorCodeCaptchaInvalid)
	}

	if hasCaptcha {
		if verifyErr := s.captchaService.Verify(cmd.CaptchaId, cmd.CaptchaAnswer); verifyErr != nil {
			dto.NeedCaptcha = true
			return domain.NewDomainError(domain.ErrorCodeCaptchaInvalid)
		}
	}
	return nil
}

func (s *userService) doLogin(cmd *CmdToLogin, lid string) (domain.User, domain.Login, error) {
	if cmd.Account != nil {
		return s.ls.LoginByAccount(cmd.LinkId, cmd.Account, cmd.Password)
	}
	return s.ls.LoginByEmail(cmd.LinkId, cmd.Email, cmd.Password)
}

func (s *userService) clearLoginFailure(lid string) {
	if err := s.ls.ClearLoginFailure(lid); err != nil {
		logs.Warn("clear login failure failed, err: %s", err.Error())
	}
}

func (s *userService) fillLoginResult(cmd *CmdToLogin, dto *UserLoginDTO, u domain.User) error {
	if err := s.checkPrivacyConsent(cmd.PrivacyConsented, &u); err != nil {
		return err
	}

	role, err := s.getRole(&u)
	if err != nil {
		return err
	}

	dto.Role = role
	dto.Email = u.EmailAddr.EmailAddr()
	dto.UserId = u.Id
	dto.CorpSigningId = u.CorpSigningId
	dto.PrivacyVersion = u.PrivacyConsent.Version
	dto.InitialPWChanged = u.PasswordChanged

	return nil
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
