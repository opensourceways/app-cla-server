package domain

import (
	"errors"
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type Manager struct {
	Id string
	Representative
}

func (m *Manager) isEmpty() bool {
	return m.Id == ""
}

func (m *Manager) Account() (dp.Account, error) {
	if m.isEmpty() {
		return nil, errors.New("not a manager")
	}

	return dp.NewAccount(fmt.Sprintf("%s_%s", m.Id, m.EmailAddr.Domain()))
}
