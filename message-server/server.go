package main

import (
	"context"
	"encoding/json"

	kafka "github.com/opensourceways/kafka-lib/agent"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/infrastructure/messageimpl"
	"github.com/opensourceways/app-cla-server/models"
)

type server struct {
	service app.SoftwarePkgMessageService
}

func (s *server) run(ctx context.Context, cfg *Config) error {
	if err := s.subscribe(cfg); err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (s *server) subscribe(cfg *Config) error {
	topics := &cfg.Topics

	h := map[string]kafka.Handler{
		topics.NewSignedCorpCLA: s.handleNewSignedCorpCLA,
	}

	return kafka.Subscriber().Subscribe(cfg.GroupName, h)
}

func (s *server) handleNewSignedCorpCLA(data []byte) error {
	e, err := messageimpl.UnmarshalToNewSignedCorpCLA(data)
	if err != nil {
		return err
	}

	index := models.SigningIndex{
		LinkId:    e.LinkId,
		SigningId: e.SigningId,
	}

	// step1. check the corp cla
	if err := s.checkCorpCLA(&index); err != nil {

	}
	// step2. list all individual clas
	v, err := models.ListIndividualSigning(index.LinkId, dbmodels.IndividualSigningListOpt{
		Email: e.Email,
	})

	if err != nil || len(v) == 0 {
		return err
	}

	// step3. backup the individual cla, delete it and send email one by one
	return nil
}

func (s *server) checkCorpCLA(index *models.SigningIndex) error {
	_, err := models.GetCorpSigningBasicInfo(index)

	if err.IsErrorOf(models.ErrUnsigned) {
		return nil
	}

	return err
}

func (s *server) backupIndividuals() {}
