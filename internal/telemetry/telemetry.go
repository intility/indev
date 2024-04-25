package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adrg/xdg"
	"github.com/denisbrodbeck/machineid"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/intility/minctl/internal/build"
	"github.com/intility/minctl/internal/env"
	"github.com/intility/minctl/internal/redact"
)

type EventName int

const (
	permissionStateFile = 0o600
	permissionStateDir  = 0o700

	sentryFlushTimeout = time.Second * 3

	deviceSalt = "43381f8c-da93-4098-afde-cbf6520f492f"
	uidSalt    = "454a83fc-848f-450c-b3df-b1cfdb26e12a"
	userPrefix = "user:"
)

var (
	deviceID string
	userID   string

	// procStartTime records the start time of the current process.
	procStartTime = time.Now()
	needsFlush    atomic.Bool
	started       bool
)

// Start enables telemetry for the current program.
func Start() {
	if started || env.DoNotTrack() || build.SentryDSN == "" {
		return
	}

	deviceID, _ = machineid.ProtectedID(deviceSalt)

	username := env.Username()
	if username != "" {
		// userID is a v5 UUID which is basically a SHA hash of the username.
		// See https://www.uuidtools.com/uuid-versions-explained for a comparison of UUIDs.
		userID = uuid.NewSHA1(uuid.MustParse(uidSalt), []byte(userPrefix+username)).String()
	}

	started = true
}

// Stop stops gathering telemetry and flushes buffered events to disk.
func Stop() {
	if !started || !needsFlush.Load() {
		return
	}

	// Report errors in a separate process, so we don't block exiting.
	exe, err := os.Executable()
	if err == nil {
		_ = exec.Command(exe, "upload-telemetry").Start()
	}

	started = false
}

// Error reports an error to the telemetry server.
func Error(err error, meta Metadata) {
	errToLog := err // use errToLog to avoid shadowing err later. Use err to keep API clean.
	if !started || errToLog == nil {
		return
	}

	osInfo := build.OS()

	event := &sentry.Event{ //nolint:exhaustruct
		EventID: sentry.EventID(ExecutionID),
		Level:   sentry.LevelError,
		User: sentry.User{ //nolint:exhaustruct
			ID: deviceID,
		},
		Exception: newSentryException(redact.Error(errToLog)),
		Contexts: map[string]map[string]any{
			"app": {
				"app_identifier": "com.intility." + build.AppName,
				"app_name":       build.AppName,
				"app_version":    build.Version,
				"app_build":      build.Commit,
				"app_start_time": procStartTime.Format(time.RFC3339),
			},
			"os": {
				// "version":        osInfo.Core,
				"name":           osInfo.Name,
				"kernel_version": osInfo.Core,
				"rooted":         osInfo.Rooted,
			},
			"device": {
				"arch": runtime.GOARCH,
			},
			"runtime": {
				"name":    "Go",
				"version": strings.TrimPrefix(runtime.Version(), "go"),
			},
		},
	}

	event.Tags = meta.cmdTags()

	if sentryCtx := meta.cmdContext(); len(sentryCtx) > 0 {
		event.Contexts["Command"] = sentryCtx
	}

	if sentryCtx := meta.envContext(); len(sentryCtx) > 0 {
		event.Contexts["CLI Environment"] = sentryCtx
	}

	// Prefer using the user ID instead of the device ID when it's
	// available.
	if userID != "" {
		event.User.ID = userID
	}

	bufferSentryEvent(event)
}

type Metadata struct {
	Command      string
	CommandFlags []string

	CustomProperty string
}

func (m *Metadata) cmdContext() map[string]any {
	sentryCtx := map[string]any{}
	if m.Command != "" {
		sentryCtx["Name"] = m.Command
	}

	if len(m.CommandFlags) > 0 {
		sentryCtx["Flags"] = m.CommandFlags
	}

	if !procStartTime.IsZero() {
		sentryCtx["DurationMS"] = time.Since(procStartTime).Milliseconds()
	}

	return sentryCtx
}

func (m *Metadata) envContext() map[string]any {
	sentryCtx := map[string]any{
		"Custom Property": m.CustomProperty,
		"Shell":           env.UserShell(),
	}

	return sentryCtx
}

func (m *Metadata) cmdTags() map[string]string {
	tags := map[string]string{}
	if m.Command != "" {
		tags["command"] = m.Command
	}

	performance := "normal"

	if time.Since(procStartTime) > time.Second {
		performance = "slow"
	}

	tags["performance"] = performance

	return tags
}

var sentryBufferDir = filepath.Join(xdg.StateHome, "minctl", "sentry")

func Upload() {
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		if !initSentryClient(build.AppName) {
			return
		}

		events := restoreEvents[sentry.Event](sentryBufferDir)

		for _, event := range events {
			// implicit memory aliasing is not a thing in Go 1.22
			// as the iteration variable is redeclared in each iteration.
			sentry.CaptureEvent(&event) //nolint:gosec
		}

		sentry.Flush(sentryFlushTimeout)
	}()

	wg.Wait()
}

func restoreEvents[E any](dir string) []E {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	events := make([]E, 0, len(dirEntries))

	for _, entry := range dirEntries {
		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)

		// Always delete the file, so we don't end up with an infinitely growing
		// backlog of errors.
		_ = os.Remove(path)

		if err != nil {
			continue
		}

		var event E
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		events = append(events, event)
	}

	return events
}

func bufferEvent(file string, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	err = os.WriteFile(file, data, permissionStateFile)
	if errors.Is(err, fs.ErrNotExist) {
		// XDG specifies perms 0700.
		if err := os.MkdirAll(filepath.Dir(file), permissionStateDir); err != nil {
			return
		}

		err = os.WriteFile(file, data, permissionStateFile)
	}

	if err == nil {
		needsFlush.Store(true)
	}
}

func newEventID() string {
	// Generate event UUIDs the same way the Sentry SDK does:
	// https://github.com/getsentry/sentry-go/blob/d9ce5344e7e1819921ea4901dd31e47a200de7e0/util.go#L15
	idLength := 16
	id := make([]byte, idLength)
	_, _ = rand.Read(id)
	id[6] &= 0x0F
	id[6] |= 0x40
	id[8] &= 0x3F
	id[8] |= 0x80

	return hex.EncodeToString(id)
}
