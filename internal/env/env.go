package env

import (
	"os"
	"os/user"
	"strconv"
)

const (
	eDoNotTrack = "DO_NOT_TRACK"
)

// system.
const (
	Env   = "ENV"
	Home  = "HOME"
	Path  = "PATH"
	Shell = "SHELL"
	User  = "USER"
)

func DoNotTrack() bool {
	// https://consoledonottrack.com/
	doNotTrack, _ := strconv.ParseBool(os.Getenv(eDoNotTrack))
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

func UserShell() string {
	return os.Getenv(Shell)
}
