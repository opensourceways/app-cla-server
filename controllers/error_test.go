package controllers

import (
	"testing"

	"github.com/opensourceways/app-cla-server/models"
)

func TestParseModelError(t *testing.T) {
	if r := parseModelError(nil); r != nil {
		t.Error("expected nil for nil input")
	}

	err := models.NewModelError(models.ErrSystemError, nil)
	r := parseModelError(err)
	if r.statusCode != 500 {
		t.Errorf("expected 500, got %d", r.statusCode)
	}

	err = models.NewModelError(models.ErrUnknownDBError, nil)
	r = parseModelError(err)
	if r.statusCode != 500 {
		t.Errorf("expected 500, got %d", r.statusCode)
	}

	err = models.NewModelError(models.ErrNoPermission, nil)
	r = parseModelError(err)
	if r.statusCode != 401 {
		t.Errorf("expected 401, got %d", r.statusCode)
	}

	err = models.NewModelError(models.ErrUnsigned, nil)
	r = parseModelError(err)
	if r.statusCode != 400 {
		t.Errorf("expected 400, got %d", r.statusCode)
	}
}

func TestParseModelErrorDefaultCode(t *testing.T) {
	err := models.NewModelError("unknown_error_code", nil)
	r := parseModelError(err)
	if r.statusCode != 400 {
		t.Errorf("expected 400, got %d", r.statusCode)
	}
}

func TestFetchInputPayloadData(t *testing.T) {
	input := []byte(`{"name": "test"}`)
	var result map[string]interface{}
	if fr := fetchInputPayloadData(input, &result); fr != nil {
		t.Errorf("unexpected error: %+v", fr)
	}
	if result["name"] != "test" {
		t.Errorf("expected 'test', got '%v'", result["name"])
	}
}

func TestFetchInputPayloadDataInvalid(t *testing.T) {
	input := []byte(`{invalid json`)
	var result map[string]interface{}
	if fr := fetchInputPayloadData(input, &result); fr == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNewFailedApiResult(t *testing.T) {
	fr := newFailedApiResult(400, "test_code", nil)
	if fr == nil {
		t.Fatal("expected non-nil")
	}
	if fr.statusCode != 400 {
		t.Errorf("expected 400, got %d", fr.statusCode)
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []string{
		errSystemError, errMissingToken, errUnknownToken, errExpiredToken,
		errUnauthorizedToken, errMissingURLPathParameter, errReadingFile,
		errParsingApiBody, errResigned, errUnsigned, errNoLink,
		errWrongIDOrPassword, errLinkExists, errCLAExists, errCLAIsUsed,
		errAuthFailed, errUnsupportedCodePlatform, errUnsupportedEmailPlatform,
		errNotYoursOrg, errNoCorpEmployeeManager, errGoToSignEmployeeCLA,
	}
	for _, c := range codes {
		if c == "" {
			t.Error("empty error code")
		}
	}
}
