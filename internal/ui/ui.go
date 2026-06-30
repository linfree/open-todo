package ui

import (
	"fmt"
	"os/exec"
	"runtime"
)

type UI interface {
	Run(onReady func())
	Quit()
}

func New(port int) UI {
	return newUI(port)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

func localURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}
