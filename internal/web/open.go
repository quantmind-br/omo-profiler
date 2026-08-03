package web

import (
	"os/exec"
	"runtime"
)

// openBrowser opens url in the platform's default browser. Errors are the
// caller's to ignore (headless / SSH environments have no browser).
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux, bsd, etc.
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
