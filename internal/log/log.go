// Package log provides leveled logging with real-time WebSocket broadcast.
package log

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Entry is a single log line.
type Entry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"msg"`
}

var (
	mu   sync.RWMutex
	subs = make(map[chan Entry]struct{})
	buf  []Entry
)

const maxBuf = 500

// Subscribe returns a channel receiving new log entries.
func Subscribe() (chan Entry, func()) {
	ch := make(chan Entry, 100)
	mu.Lock()
	subs[ch] = struct{}{}
	// replay recent
	for _, e := range buf {
		select {
		case ch <- e:
		default:
		}
	}
	mu.Unlock()
	return ch, func() {
		mu.Lock()
		delete(subs, ch)
		close(ch)
		mu.Unlock()
	}
}

func emit(level, format string, args ...any) {
	e := Entry{
		Time:    time.Now().Format("15:04:05"),
		Level:   level,
		Message: fmt.Sprintf(format, args...),
	}
	mu.Lock()
	buf = append(buf, e)
	if len(buf) > maxBuf {
		buf = buf[len(buf)-maxBuf:]
	}
	for ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
	mu.Unlock()
	if level == "ERROR" {
		log.Printf("[ERROR] %s", e.Message)
	}
}

// Info logs at info level.
func Info(format string, args ...any) { emit("INFO", format, args...) }

// Warn logs at warn level.
func Warn(format string, args ...any) { emit("WARN", format, args...) }

// Error logs at error level.
func Error(format string, args ...any) { emit("ERROR", format, args...) }
