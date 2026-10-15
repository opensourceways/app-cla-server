package watch

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSendCLAUpdatedEventDroppedLogsError verifies that when the event channel
// is full (or has no reader), SendCLAUpdatedEvent hits the default branch and
// emits a structured error log whose business kv uses "event_msg" (not "msg",
// which is the obs-sdk contract field for the log message phrase).
func TestSendCLAUpdatedEventDroppedLogsError(t *testing.T) {
	orig := claUpdatedWatchInstance
	defer func() { claUpdatedWatchInstance = orig }()

	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		genCLADiff: make(chan message.CLAUpdatedMsg, 0),
	}

	buf, cleanup := initWatchTestLogger(t)
	defer cleanup()

	msg := message.CLAUpdatedMsg{
		LinkId:   "link1",
		OldCLAId: "cla-old",
		NewCLAId: "cla-new",
	}
	SendCLAUpdatedEvent(msg)

	entries := parseWatchLogLines(t, buf)
	require.Len(t, entries, 1)
	e := entries[0]

	assert.Equal(t, "error", e["level"])
	assert.Equal(t, "send cla updated event failed", e["msg"])

	em, ok := e["event_msg"]
	require.True(t, ok, "log should contain event_msg business kv key")
	assert.NotNil(t, em)
}
