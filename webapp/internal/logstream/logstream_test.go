package logstream

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewStreamer(t *testing.T) {
	streamer := NewStreamer(2, 1000)
	if streamer == nil {
		t.Fatal("expected non-nil streamer")
	}
	if streamer.pollInterval != 2*time.Second {
		t.Errorf("expected pollInterval 2s, got %v", streamer.pollInterval)
	}
	if streamer.maxLines != 1000 {
		t.Errorf("expected maxLines 1000, got %d", streamer.maxLines)
	}
}

func TestParseLogLine_ValidFormat(t *testing.T) {
	line := "Jan 02 15:04:05 hostname factorio@instance[123]: Test log message"
	entry := parseLogLine(line)

	if entry.Message == "" {
		t.Error("expected non-empty message")
	}
	if !strings.Contains(entry.Message, "factorio@instance") {
		t.Errorf("expected 'factorio@instance' in message, got %s", entry.Message)
	}
	if !strings.Contains(entry.Message, "Test log message") {
		t.Errorf("expected 'Test log message' in message, got %s", entry.Message)
	}
}

func TestParseLogLine_InvalidFormat(t *testing.T) {
	line := "Invalid log line without timestamp"
	entry := parseLogLine(line)

	if entry.Message == "" {
		t.Error("expected non-empty message for invalid format")
	}
	if entry.Message != line {
		t.Errorf("expected message to be the entire line, got %s", entry.Message)
	}
}

func TestParseLogLine_ShortFormat(t *testing.T) {
	line := "Short"
	entry := parseLogLine(line)

	if entry.Message == "" {
		t.Error("expected non-empty message for short line")
	}
	if entry.Message != line {
		t.Errorf("expected message to be the entire line, got %s", entry.Message)
	}
}

func TestParseLogOutput_ValidOutput(t *testing.T) {
	output := `Jan 02 15:04:05 hostname factorio@instance[123]: First line
Jan 02 15:04:06 hostname factorio@instance[123]: Second line

Jan 02 15:04:07 hostname factorio@instance[123]: Third line
`

	entries := parseLogOutput(output)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestParseLogOutput_EmptyOutput(t *testing.T) {
	entries := parseLogOutput("")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty output, got %d", len(entries))
	}
}

func TestParseJSONLogLine_ValidJSON(t *testing.T) {
	line := `{"__REALTIME_TIMESTAMP":"1735689600000000","MESSAGE":"Test message"}`
	entry := parseJSONLogLine(line)

	if entry.Message != "Test message" {
		t.Errorf("expected 'Test message', got %s", entry.Message)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParseJSONLogLine_MissingFields(t *testing.T) {
	line := `{"FIELD":"value"}`
	entry := parseJSONLogLine(line)

	if entry.Message != "" {
		t.Errorf("expected empty message, got %s", entry.Message)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected current time for missing timestamp")
	}
}

func TestParseJSONLogLine_EmptyLine(t *testing.T) {
	line := ""
	entry := parseJSONLogLine(line)

	if entry.Message != "" {
		t.Errorf("expected empty message, got %s", entry.Message)
	}
}

func TestParseMicroseconds_ValidInput(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"1234567890", 1234567890},
		{"000123456", 123456},
		{"123abc", 123},
		{"", 0},
	}

	for _, tc := range testCases {
		result, err := parseMicroseconds(tc.input)
		if err != nil {
			t.Errorf("expected no error for '%s', got %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("expected %d for '%s', got %d", tc.expected, tc.input, result)
		}
	}
}

func TestStreamLogs_ContextCancellation(t *testing.T) {
	streamer := NewStreamer(1, 100)
	ctx, cancel := context.WithCancel(context.Background())

	ch := streamer.StreamLogs(ctx, "test")

	time.Sleep(50 * time.Millisecond)
	cancel()

	timeout := time.After(500 * time.Millisecond)
	select {
	case <-ch:
	case <-timeout:
		t.Fatal("expected channel to close on context cancellation")
	}
}

func TestStreamLogsFollow_ContextCancellation(t *testing.T) {
	streamer := NewStreamer(1, 100)
	ctx, cancel := context.WithCancel(context.Background())

	_, err := streamer.StreamLogsFollow(ctx, "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	time.Sleep(100 * time.Millisecond)
}

func TestGetLogs_InvalidLinesParameter(t *testing.T) {
	streamer := NewStreamer(1, 100)

	linesParamTests := []struct {
		input  int
		expect int
	}{
		{0, 100},
		{-10, 100},
		{200, 100},
		{50, 50},
	}

	for _, tc := range linesParamTests {
		entries, err := streamer.GetLogs("test", tc.input)
		if err == nil && len(entries) != 0 {
			t.Logf("Note: GetLogs called with %d lines, expected to use maxLines=%d", tc.input, tc.expect)
		}
	}
}

func TestLogEntry_Formatting(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		Timestamp: now,
		Message:   "Test message",
	}

	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if entry.Message != "Test message" {
		t.Errorf("expected 'Test message', got %s", entry.Message)
	}
}

func TestParseLogLine_Replacement(t *testing.T) {
	line := "Jan 02 15:04:05 hostname factorio factorio-instance[123]: Test"
	entry := parseLogLine(line)

	if strings.Contains(entry.Message, "factorio factorio-") {
		t.Error("expected 'factorio factorio-' to be replaced with 'factorio@'")
	}
	if !strings.Contains(entry.Message, "factorio@instance") {
		t.Errorf("expected 'factorio@instance' in message, got %s", entry.Message)
	}
}

func TestParseLogOutput_SkipsEmptyLines(t *testing.T) {
	output := `

Line 1

Line 2

`

	entries := parseLogOutput(output)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries (empty lines skipped), got %d", len(entries))
	}
}
