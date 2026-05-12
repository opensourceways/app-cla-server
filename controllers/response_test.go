package controllers

import (
	"testing"
)

func TestRespData(t *testing.T) {
	rd := respData{Data: "test"}
	if rd.Data != "test" {
		t.Errorf("expected 'test', got '%v'", rd.Data)
	}
}

func TestErrMsg(t *testing.T) {
	em := errMsg{ErrCode: "E001", ErrMsg: "error message"}
	if em.ErrCode != "E001" || em.ErrMsg != "error message" {
		t.Errorf("errMsg mismatch")
	}
}

func TestHeaderConstants(t *testing.T) {
	if headerToken != "Token" {
		t.Errorf("expected 'Token', got '%s'", headerToken)
	}
	if headerPasswordRetrievalKey != "Password-Retrieval-Key" {
		t.Errorf("expected 'Password-Retrieval-Key', got '%s'", headerPasswordRetrievalKey)
	}
	if fileNameOfUploadingOrgSignatue != "org_signature_file" {
		t.Errorf("expected 'org_signature_file', got '%s'", fileNameOfUploadingOrgSignatue)
	}
}

func TestParamConstants(t *testing.T) {
	if ParamLinkID != ":link_id" || ParamSigningID != ":signing_id" || ParamEmail != ":email" || ParamApplyTo != ":apply_to" {
		t.Error("param constant mismatch")
	}
}

func TestAPIConstants(t *testing.T) {
	if apiAccessController != "access_controller" {
		t.Errorf("expected 'access_controller', got '%s'", apiAccessController)
	}
}
