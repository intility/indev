package build

import (
	"os"
	"os/user"
	"runtime"
	"strconv"
	"sync"

	"github.com/matishsiao/goInfo"
)

var forceProd, _ = strconv.ParseBool(os.Getenv("ICPCTL_PROD"))

// Variables in this file are set via ldflags.
var (
	AppName    = "icpctl"
	Version    = "0.0.0-dev"
	Commit     = "none"
	CommitDate = "unknown"

	//goland:noinspection GoBoolExpressions
	IsDev = Version == "0.0.0-dev" && !forceProd

	PlatformBaseURI = "https://container-platform-backend.apps.hypershift.intilitycloud.com"

	// SentryDSN is injected in the build from the CI/CD pipeline.
	// It is disabled by default.
	SentryDSN = ""

	// OTELCollectorEndpoint is injected in the build from the CI/CD pipeline.
	// It is disabled by default.
	OTELCollectorEndpoint = "https://otlp-collector.apps.hypershift.intilitycloud.com"
	OTELCollectorToken    = ""

	AuthPlatformAudience = "api://adc683b8-0523-4d0e-9f99-0a8536d4c618/user_impersonation"
	AuthAuthority        = "https://login.microsoftonline.com/intility.no"
	AuthClientID         = "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c"
	AuthRedirect         = "http://localhost:42069"
)

// User-presentable names of operating systems supported by icpctl.
const (
	OSLinux  = "Linux"
	OSDarwin = "macOS"
	OSWSL    = "WSL"
)

var (
	osInfo OSInfo
	osOnce sync.Once
)

type OSInfo struct {
	Name     string
	Kernel   string
	Core     string
	Platform string
	OS       string
	Hostname string
	Rooted   bool
}

func OS() OSInfo { //nolint:cyclop
	osOnce.Do(func() {
		osInfo.Name = runtime.GOOS

		switch runtime.GOOS {
		case "linux":
			if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
				osInfo.Name = OSWSL
			}

			if _, err := os.Stat("/run/WSL"); err == nil {
				osInfo.Name = OSWSL
			}

			osInfo.Name = OSLinux
		case "darwin":
			osInfo.Name = OSDarwin
		default:
			osInfo.Name = runtime.GOOS
		}

		gi, err := goInfo.GetInfo()
		if err != nil {
			return
		}

		osInfo.Kernel = gi.Kernel
		osInfo.Core = gi.Core
		osInfo.Platform = gi.Platform
		osInfo.OS = gi.OS
		osInfo.Hostname = gi.Hostname
		osInfo.Rooted = false

		if os.Getuid() == 0 {
			osInfo.Rooted = true
		}

		usr, err := user.Current()
		if err == nil && usr.Username == "root" {
			osInfo.Rooted = true
		}

		if os.Getenv("SUDO_UID") != "" {
			osInfo.Rooted = true
		}
	})

	return osInfo
}

func PlatformAPIHost() string {
	if IsDev {
		return "http://localhost:8080"
	}

	return PlatformBaseURI
}

func ClientID() string {
	if IsDev {
		return "b65cf9b0-290c-4b44-a4b1-0b02b7752b3c"
	}

	return AuthClientID
}

func Authority() string {
	if IsDev {
		return "https://login.microsoftonline.com/intility.no"
	}

	return AuthAuthority
}

func SuccessRedirect() string {
	if IsDev {
		return "http://localhost:42069"
	}

	return AuthRedirect
}

func Scopes() []string {
	return []string{AuthPlatformAudience}
}

func NameVersionString() string {
	if IsDev {
		return AppName + " " + Version + " (dev)"
	}

	return AppName + " v" + Version + " (" + Commit + ": " + CommitDate + ")"
}
