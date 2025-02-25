package app

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
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
	vcService vcservice.VCService,
	interval time.Duration,
) UserService {
	return &userService{
		us:        us,
		ls:        ls,
		repo:      repo,
		encrypt:   encrypt,
		vcService: verificationCodeService{vcService},
		interval:  interval,
	}
}

type UserService interface {
	Get(userId string) (dto UserBasicInfoDTO, err error)
	Login(cmd *CmdToLogin) (dto UserLoginDTO, err error)
	ResetPassword(cmd *CmdToResetPassword) error
	ChangePassword(cmd *CmdToChangePassword) error
	GenKeyForPasswordRetrieval(*CmdToGenKeyForPasswordRetrieval) (string, error)
}

type userService struct {
	us        userservice.UserService
	ls        loginservice.LoginService
	repo      repository.CorpSigning
	encrypt   symmetricencryption.Encryption
	interval  time.Duration
	vcService verificationCodeService
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

func (s *userService) Login(cmd *CmdToLogin) (dto UserLoginDTO, err error) {
	defer cmd.clear()

	var u domain.User
	var l domain.Login

	if cmd.Account != nil {
		u, l, err = s.ls.LoginByAccount(cmd.LinkId, cmd.Account, cmd.Password)
	} else {
		u, l, err = s.ls.LoginByEmail(cmd.LinkId, cmd.Email, cmd.Password)
	}

	// It should record the retry number whatever if it is success or not.
	dto.RetryNum = l.RetryNum()

	if err != nil {
		return
	}

	if err = s.checkPrivacy(false, &u); err != nil {
		return
	}

	cs, err := s.repo.Find(u.CorpSigningId)
	if err != nil {
		return
	}

	if dto.Role = cs.GetRole(u.EmailAddr); dto.Role == "" {
		err = domain.NewDomainError(domain.ErrorCodeUserWrongAccountOrPassword)

		s.us.Remove([]string{u.Id})

		return
	}

	dto.Email = u.EmailAddr.EmailAddr()
	dto.UserId = u.Id
	dto.CorpName = cs.CorpName().CorpName()
	dto.CorpSigningId = u.CorpSigningId
	dto.PrivacyConsent = u.PrivacyConsent.Version
	dto.InitialPWChanged = u.PasswordChanged

	return
}

func (s *userService) checkPrivacy(privacyConsent bool, u *domain.User) error {
	if !u.UpdatePrivacyConsent(s.privacyVersion) {
		return nil
	}

	if !privacyConsent {
		return nil //TODO
	}

	// TODO save

	return nil
}

func (s *userService) Get(userId string) (dto UserBasicInfoDTO, err error) {
	u, err := s.us.Get(userId)
	if err != nil {
		return
	}

	dto.UserId = u.Account.Account()
	dto.InitialPWChanged = u.PasswordChanged

	cs, err := s.repo.Find(u.CorpSigningId)
	if err != nil {
		return
	}

	if dto.Role = cs.GetRole(u.EmailAddr); dto.Role == "" {
		err = errors.New("no role")
	}

	return
}
