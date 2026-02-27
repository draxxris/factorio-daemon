package logstream

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// LogEntry represents a single log line
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// Streamer handles journald log streaming
type Streamer struct {
	pollInterval time.Duration
	maxLines     int
}

// NewStreamer creates a new log streamer
func NewStreamer(pollIntervalSeconds, maxLines int) *Streamer {
	return &Streamer{
		pollInterval: time.Duration(pollIntervalSeconds) * time.Second,
		maxLines:     maxLines,
	}
}

// GetLogs retrieves recent logs for an instance
func (s *Streamer) GetLogs(instance string, lines int) ([]LogEntry, error) {
	if lines <= 0 || lines > s.maxLines {
		lines = s.maxLines
	}

	cmd := exec.Command("journalctl", "-u", "factorio@"+instance, "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return parseLogOutput(string(output)), nil
}

// StreamLogs streams logs via a channel using polling
func (s *Streamer) StreamLogs(ctx context.Context, instance string) <-chan LogEntry {
	ch := make(chan LogEntry, 100)
	
	go func() {
		defer close(ch)
		
		var lastTimestamp time.Time
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			
			// Get recent logs
			entries, err := s.GetLogs(instance, 50)
			if err != nil {
				time.Sleep(s.pollInterval)
				continue
			}
			
			// Send only new entries
			for _, entry := range entries {
				if entry.Timestamp.After(lastTimestamp) {
					ch <- entry
					lastTimestamp = entry.Timestamp
				}
			}
			
			time.Sleep(s.pollInterval)
		}
	}()
	
	return ch
}

// StreamLogsFollow uses journalctl -f for real-time streaming
func (s *Streamer) StreamLogsFollow(ctx context.Context, instance string) (<-chan LogEntry, error) {
	ch := make(chan LogEntry, 100)
	
	cmd := exec.CommandContext(ctx, "journalctl", "-u", "factorio@"+instance, "-f", "--output=json")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start journalctl: %w", err)
	}
	
	go func() {
		defer close(ch)
		defer cmd.Wait()
		
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				return
			default:
			}
			
			line := scanner.Text()
			entry := parseJSONLogLine(line)
			if entry.Message != "" {
				ch <- entry
			}
		}
	}()
	
	return ch, nil
}

// parseLogOutput parses journalctl text output into LogEntry slice
func parseLogOutput(output string) []LogEntry {
	var entries []LogEntry
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		entry := parseLogLine(line)
		if entry.Message != "" {
			entries = append(entries, entry)
		}
	}
	
	return entries
}

// parseLogLine parses a single journalctl log line
// Format: "Jan 02 15:04:05 hostname factorio-instance[pid]: message"
func parseLogLine(line string) LogEntry {
	entry := LogEntry{}
	
	// Try to parse the timestamp from the beginning of the line
	// journalctl default format: "Jan 02 15:04:05 ..."
	if len(line) > 15 {
		// Extract timestamp portion
		tsPart := line[:15]
		// Parse with current year (journalctl doesn't include year)
		ts, err := time.ParseInLocation("Jan 02 15:04:05", tsPart, time.Local)
		if err == nil {
			// Set year to current year
			now := time.Now()
			ts = ts.AddDate(now.Year(), 0, 0)
			entry.Timestamp = ts
			message := strings.TrimSpace(line[15:])
			// Replace "factorio factorio-XXX" with "factorio@XXX"
			message = strings.Replace(message, "factorio factorio-", "factorio@", 1)
			entry.Message = message
		} else {
			// Fallback: use current time and whole line as message
			entry.Timestamp = time.Now()
			entry.Message = line
		}
	} else {
		entry.Timestamp = time.Now()
		entry.Message = line
	}
	
	return entry
}

// parseJSONLogLine parses a JSON log line from journalctl --output=json
func parseJSONLogLine(line string) LogEntry {
	entry := LogEntry{}
	
	// Simple JSON parsing for the fields we need
	// Look for __REALTIME_TIMESTAMP and MESSAGE fields
	
	// Extract timestamp
	if tsIdx := strings.Index(line, `"__REALTIME_TIMESTAMP":"`); tsIdx != -1 {
		start := tsIdx + len(`"__REALTIME_TIMESTAMP":"`)
		end := strings.Index(line[start:], `"`)
		if end != -1 {
			tsStr := line[start : start+end]
			// Timestamp is in microseconds
			if tsMicro, err := parseMicroseconds(tsStr); err == nil {
				entry.Timestamp = time.UnixMicro(tsMicro)
			}
		}
	}
	
	// Extract message
	if msgIdx := strings.Index(line, `"MESSAGE":"`); msgIdx != -1 {
		start := msgIdx + len(`"MESSAGE":"`)
		end := strings.Index(line[start:], `"`)
		if end != -1 {
			entry.Message = line[start : start+end]
		}
	}
	
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	
	return entry
}

// parseMicroseconds parses a microsecond timestamp string
func parseMicroseconds(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else {
			break
		}
	}
	return result, nil
}
