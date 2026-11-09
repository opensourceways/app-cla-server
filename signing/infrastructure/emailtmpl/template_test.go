package emailtmpl

import (
	"strings"
	"testing"

	"github.com/opensourceways/app-cla-server/util"
)

// TestCorporationSigningTemplateRendering verifies that the
// corporation-signing.tmpl template, when rendered with a
// CorporationSigning struct containing PDFSizeLimit, produces output
// that includes the PDF size limit, the name consistency requirement,
// and the overall signing flow steps.
func TestCorporationSigningTemplateRendering(t *testing.T) {
	tmpl, err := util.NewTemplate(
		"corporation signing test",
		"../../../conf/email-template/corporation-signing.tmpl",
	)
	if err != nil {
		t.Fatalf("failed to load template: %s", err)
	}

	data := CorporationSigning{
		Org:          "test-org",
		Date:         "2024-01-01",
		AdminName:    "admin",
		ProjectURL:   "http://project.example.com",
		SigningInfo:  "Corporation: Test Corp",
		PDFSizeLimit: "5 MB",
	}

	buf, err := util.RenderTemplate(tmpl, &data)
	if err != nil {
		t.Fatalf("failed to render template: %s", err)
	}

	output := buf.String()

	if !strings.Contains(output, "5 MB") {
		t.Errorf("rendered template should contain PDF size limit '5 MB'")
	}
	if !strings.Contains(output, "must match the legal entity name") {
		t.Errorf("rendered template should contain name consistency requirement")
	}
	if !strings.Contains(output, "Download the CLA PDF document") {
		t.Errorf("rendered template should contain signing flow step about downloading PDF")
	}
	if !strings.Contains(output, "Reply to this email with the stamped PDF") {
		t.Errorf("rendered template should contain signing flow step about replying")
	}
	if !strings.Contains(output, "signing process is completed") {
		t.Errorf("rendered template should contain signing completion step")
	}
}
