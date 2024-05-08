package env

import (
	"os"
	"os/user"
	"strconv"

	"github.com/intility/icpctl/internal/build"
)

const (
	envKeyDoNotTrack = "DO_NOT_TRACK"
)

// system.
const (
	Shell = "SHELL"
	User  = "USER"
)

func DoNotTrack() bool {
	// https://consoledonottrack.com/
	doNotTrack, _ := strconv.ParseBool(os.Getenv(envKeyDoNotTrack))
	return doNotTrack
}

func Username() string {
	usr, err := user.Current()
	if err == nil {
		return usr.Username
	}

	username := os.Getenv(User)
	if username != "" {
		return username
	}

	return "NN"
}

func OtelExporterEndpoint() string {
	return build.OTELCollectorEndpoint
}

func OtelExporterToken() string {
	return build.OTELCollectorToken
}

func SentryDSN() string {
	return build.SentryDSN
}

func UserShell() string {
	return os.Getenv(Shell)
}
