/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mattn/go-isatty"
	"sigs.k8s.io/kind/pkg/log"
)

// a fake TTY type for testing that can only be implemented within this package
type isTestFakeTTY interface {
	isTestFakeTTY()
}

// IsTerminal returns true if the writer w is a terminal
func IsTerminal(w io.Writer) bool {
	// check for internal fake type we can use for testing.
	if _, ok := (w).(isTestFakeTTY); ok {
		return true
	}
	// check for real terminals
	if v, ok := (w).(*os.File); ok {
		return isatty.IsTerminal(v.Fd())
	}
	return false
}

func isSmartTerminal(w io.Writer, GOOS string, lookupEnv func(string) (string, bool)) bool {
	// Not smart if it's not a tty
	if !IsTerminal(w) {
		return false
	}

	// getenv helper for when we only care about the value
	getenv := func(e string) string {
		v, _ := lookupEnv(e)
		return v
	}

	// Explicit request for no ANSI escape codes
	// https://no-color.org/
	if _, set := lookupEnv("NO_COLOR"); set {
		return false
	}

	// Explicitly dumb terminals are not smart
	// https://en.wikipedia.org/wiki/Computer_terminal#Dumb_terminals
	term := getenv("TERM")
	if term == "dumb" {
		return false
	}
	// st has some bug 🤷‍♂️
	// https://github.com/kubernetes-sigs/kind/issues/1892
	if term == "st-256color" {
		return false
	}

	// On Windows WT_SESSION is set by the modern terminal component.
	// Older terminals have poor support for UTF-8, VT escape codes, etc.
	if GOOS == "windows" && getenv("WT_SESSION") == "" {
		return false
	}

	/* CI Systems with bad Fake TTYs */
	// Travis CI
	// https://github.com/kubernetes-sigs/kind/issues/1478
	// We can detect it with documented magical environment variables
	// https://docs.travis-ci.com/user/environment-variables/#default-environment-variables
	if getenv("HAS_JOSH_K_SEAL_OF_APPROVAL") == "true" && getenv("TRAVIS") == "true" {
		return false
	}

	// OK, we'll assume it's smart now, given no evidence otherwise.
	return true
}

// trivial fake TTY writer for testing
type testFakeTTY struct{}

func (t *testFakeTTY) Write(p []byte) (int, error) {
	return len(p), nil
}

func (t *testFakeTTY) isTestFakeTTY() {}

// IsSmartTerminal returns true if the writer w is a terminal AND
// we think that the terminal is smart enough to use VT escape codes etc.
func IsSmartTerminal(w io.Writer) bool {
	return isSmartTerminal(w, runtime.GOOS, os.LookupEnv)
}

// Logger is the kind cli's log.Logger implementation
type Logger struct {
	writer     io.Writer
	writerMu   sync.Mutex
	verbosity  log.Level
	bufferPool *bufferPool
	// kind special additions
	isSmartWriter bool
}

var _ log.Logger = &Logger{}

// NewLogger returns a new Logger with the given verbosity
func NewLogger(writer io.Writer, verbosity log.Level) *Logger {
	l := &Logger{
		verbosity:  verbosity,
		bufferPool: newBufferPool(),
	}
	l.SetWriter(writer)
	return l
}

// SetWriter sets the output writer
func (l *Logger) SetWriter(w io.Writer) {
	l.writerMu.Lock()
	defer l.writerMu.Unlock()
	l.writer = w
	_, isSpinner := w.(*Spinner)
	l.isSmartWriter = isSpinner || IsSmartTerminal(w)
}

// ColorEnabled returns true if the caller is OK to write colored output
func (l *Logger) ColorEnabled() bool {
	l.writerMu.Lock()
	defer l.writerMu.Unlock()
	return l.isSmartWriter
}

func (l *Logger) getVerbosity() log.Level {
	return log.Level(atomic.LoadInt32((*int32)(&l.verbosity)))
}

// SetVerbosity sets the loggers verbosity
func (l *Logger) SetVerbosity(verbosity log.Level) {
	atomic.StoreInt32((*int32)(&l.verbosity), int32(verbosity))
}

// synchronized write to the inner writer
func (l *Logger) write(p []byte) (n int, err error) {
	l.writerMu.Lock()
	defer l.writerMu.Unlock()
	return l.writer.Write(p)
}

// writeBuffer writes buf with write, ensuring there is a trailing newline
func (l *Logger) writeBuffer(buf *bytes.Buffer) {
	// ensure trailing newline
	if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	// TODO: should we handle this somehow??
	// Who logs for the logger? 🤔
	_, _ = l.write(buf.Bytes())
}

// print writes a simple string to the log writer
func (l *Logger) print(message string) {
	buf := bytes.NewBufferString(message)
	l.writeBuffer(buf)
}

// printf is roughly fmt.Fprintf against the log writer
func (l *Logger) printf(format string, args ...interface{}) {
	buf := l.bufferPool.Get()
	fmt.Fprintf(buf, format, args...)
	l.writeBuffer(buf)
	l.bufferPool.Put(buf)
}

// addDebugHeader inserts the debug line header to buf
func addDebugHeader(buf *bytes.Buffer) {
	_, file, line, ok := runtime.Caller(3)
	// lifted from klog
	if !ok {
		file = "???"
		line = 1
	} else {
		if slash := strings.LastIndex(file, "/"); slash >= 0 {
			path := file
			file = path[slash+1:]
			if dirsep := strings.LastIndex(path[:slash], "/"); dirsep >= 0 {
				file = path[dirsep+1:]
			}
		}
	}
	buf.Grow(len(file) + 11) // we know at least this many bytes are needed
	buf.WriteString("DEBUG: ")
	buf.WriteString(file)
	buf.WriteByte(':')
	fmt.Fprintf(buf, "%d", line)
	buf.WriteByte(']')
	buf.WriteByte(' ')
}

// debug is like print but with a debug log header
func (l *Logger) debug(message string) {
	buf := l.bufferPool.Get()
	addDebugHeader(buf)
	buf.WriteString(message)
	l.writeBuffer(buf)
	l.bufferPool.Put(buf)
}

// debugf is like printf but with a debug log header
func (l *Logger) debugf(format string, args ...interface{}) {
	buf := l.bufferPool.Get()
	addDebugHeader(buf)
	fmt.Fprintf(buf, format, args...)
	l.writeBuffer(buf)
	l.bufferPool.Put(buf)
}

// Warn is part of the log.Logger interface
func (l *Logger) Warn(message string) {
	l.print(message)
}

// Warnf is part of the log.Logger interface
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.printf(format, args...)
}

// Error is part of the log.Logger interface
func (l *Logger) Error(message string) {
	l.print(message)
}

// Errorf is part of the log.Logger interface
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.printf(format, args...)
}

// V is part of the log.Logger interface
func (l *Logger) V(level log.Level) log.InfoLogger {
	return infoLogger{
		logger:  l,
		level:   level,
		enabled: level <= l.getVerbosity(),
	}
}

// infoLogger implements log.InfoLogger for Logger
type infoLogger struct {
	logger  *Logger
	level   log.Level
	enabled bool
}

// Enabled is part of the log.InfoLogger interface
func (i infoLogger) Enabled() bool {
	return i.enabled
}

// Info is part of the log.InfoLogger interface
func (i infoLogger) Info(message string) {
	if !i.enabled {
		return
	}
	// for > 0, we are writing debug messages, include extra info
	if i.level > 0 {
		i.logger.debug(message)
	} else {
		i.logger.print(message)
	}
}

// Infof is part of the log.InfoLogger interface
func (i infoLogger) Infof(format string, args ...interface{}) {
	if !i.enabled {
		return
	}
	// for > 0, we are writing debug messages, include extra info
	if i.level > 0 {
		i.logger.debugf(format, args...)
	} else {
		i.logger.printf(format, args...)
	}
}

// bufferPool is a type safe sync.Pool of *byte.Buffer, guaranteed to be Reset
type bufferPool struct {
	sync.Pool
}

// newBufferPool returns a new bufferPool
func newBufferPool() *bufferPool {
	return &bufferPool{
		sync.Pool{
			New: func() interface{} {
				// The Pool's New function should generally only return pointer
				// types, since a pointer can be put into the return interface
				// value without an allocation:
				return new(bytes.Buffer)
			},
		},
	}
}

// Get obtains a buffer from the pool
func (b *bufferPool) Get() *bytes.Buffer {
	return b.Pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool, resetting it first
func (b *bufferPool) Put(x *bytes.Buffer) {
	// only store small buffers to avoid pointless allocation
	// avoid keeping arbitrarily large buffers
	if x.Len() > 256 {
		return
	}
	x.Reset()
	b.Pool.Put(x)
}
