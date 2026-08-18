package claconfirmtokenimpl

import (
	"encoding/hex"
	"regexp"
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randombytesimpl"
)

const tokenBytes = 32 // 256-bit random token, exported as 64 hex chars

var tokenPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func NewCLAConfirmTokenImpl(d dao) *claConfirmTokenImpl {
	return &claConfirmTokenImpl{dao: d}
}

type claConfirmTokenImpl struct {
	dao dao
}

func (impl *claConfirmTokenImpl) Add(linkId, email, newClaId string, ttl time.Duration) (string, error) {
	token, err := genToken()
	if err != nil {
		return "", err
	}

	payload := claConfirmTokenDO{
		LinkId:   linkId,
		Email:    email,
		NewCLAId: newClaId,
		IssuedAt: time.Now().Unix(),
	}

	// The index key guarantees at most one valid token per (linkId, email):
	// generating a new token invalidates the previous one.
	idxKey := indexKey(linkId, email)
	if old, err := impl.getTokenFromIndex(idxKey); err != nil {
		return "", err
	} else if old != "" {
		if err := impl.dao.Del(tokenKey(old)); err != nil {
			return "", err
		}
	}

	if err := impl.dao.SetWithExpiry(tokenKey(token), &payload, ttl); err != nil {
		return "", err
	}

	if err := impl.dao.SetWithExpiry(idxKey, token, ttl); err != nil {
		_ = impl.dao.Del(tokenKey(token))
		return "", err
	}

	return token, nil
}

func (impl *claConfirmTokenImpl) Consume(token string) (repository.CLAConfirmTokenPayload, error) {
	if !tokenPattern.MatchString(token) {
		return repository.CLAConfirmTokenPayload{}, domain.NewDomainError(domain.ErrorCodeCLAConfirmTokenInvalid)
	}

	var payload claConfirmTokenDO
	if err := impl.dao.GetAndDelete(tokenKey(token), &payload); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return repository.CLAConfirmTokenPayload{}, commonRepo.NewErrorResourceNotFound(err)
		}

		return repository.CLAConfirmTokenPayload{}, err
	}

	return repository.CLAConfirmTokenPayload{
		LinkId:   payload.LinkId,
		Email:    payload.Email,
		NewCLAId: payload.NewCLAId,
		IssuedAt: time.Unix(payload.IssuedAt, 0),
	}, nil
}

func (impl *claConfirmTokenImpl) getTokenFromIndex(idxKey string) (string, error) {
	var token string
	if err := impl.dao.Get(idxKey, &token); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return "", nil
		}

		return "", err
	}

	return token, nil
}

func genToken() (string, error) {
	b, err := randombytesimpl.NewRandomBytesImpl().New(tokenBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func tokenKey(token string) string {
	return "cla-confirm:token:" + token
}

func indexKey(linkId, email string) string {
	return "cla-confirm:idx:" + linkId + ":" + email
}
