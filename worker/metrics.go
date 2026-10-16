package worker

import (
	obsmetrics "github.com/opensourceways/obs-sdk/go/metrics"
)

var (
	emailSentCounter    *obsmetrics.CounterVec
	pdfGeneratedCounter *obsmetrics.CounterVec
)

func InitBusinessMetrics(m *obsmetrics.Metrics) {
	emailSentCounter = m.NewCounterVec(
		"cla_email_sent_total",
		"CLA notification emails sent",
		"platform", "outcome",
	)
	pdfGeneratedCounter = m.NewCounterVec(
		"cla_pdf_generated_total",
		"CLA PDF documents generated",
		"outcome",
	)
}

func recordEmailSent(platform, outcome string) {
	if emailSentCounter != nil {
		emailSentCounter.Inc(platform, outcome)
	}
}

func recordPDFGenerated(outcome string) {
	if pdfGeneratedCounter != nil {
		pdfGeneratedCounter.Inc(outcome)
	}
}
