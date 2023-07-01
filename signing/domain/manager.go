package domain

import (
	"errors"
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type Manager struct {
	Id string
	Representative

	// to avoid generate repeatly
	account dp.Account
}

func (m *Manager) IsEmpty() bool {
	return m.Id == ""
}

func (m *Manager) IsSame(m1 *Manager) bool {
	return m.EmailAddr.EmailAddr() == m1.EmailAddr.EmailAddr() || m.Id == m1.Id
}

func (m *Manager) hasEmail(e dp.EmailAddr) bool {
	return m.EmailAddr.EmailAddr() == e.EmailAddr()
}

func (m *Manager) Account() (dp.Account, error) {
	if m.account != nil {
		return m.account, nil
	}

	if m.IsEmpty() {
		return nil, errors.New("not a manager")
	}

	v, err := dp.NewAccount(fmt.Sprintf("%s_%s", m.Id, m.EmailAddr.Domain()))
	if err != nil {
		return nil, NewDomainError(ErrorCodeUserInvalidAccount)
	}

	m.account = v

	return v, nil
}
