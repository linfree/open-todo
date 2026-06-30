//go:build windows

package ui

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
	"github.com/jchv/go-webview2"
)

type windowsUI struct {
	port    int
	quitC   chan struct{}
	mu      sync.Mutex
	webview webview2.WebView
	showReq chan struct{}
	started bool
}

func newUI(port int) UI {
	return &windowsUI{
		port:    port,
		quitC:   make(chan struct{}),
		showReq: make(chan struct{}, 1),
	}
}

func (u *windowsUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		systray.SetTooltip(fmt.Sprintf("Open Todo (localhost:%d)", u.port))

		mShow := systray.AddMenuItem("显示窗口", "显示应用窗口")
		mOpen := systray.AddMenuItem("在浏览器中打开", "在外部浏览器中打开")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("退出", "退出应用")

		if onReady != nil {
			onReady()
		}

		u.startWebviewLoop()
		u.showWindow()

		url := localURL(u.port)

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					u.showWindow()
				case <-mOpen.ClickedCh:
					openBrowser(url)
				case <-mQuit.ClickedCh:
					u.closeWebview()
					systray.Quit()
					return
				case <-u.quitC:
					u.closeWebview()
					systray.Quit()
					return
				}
			}
		}()
	}, func() {
		// onExit — noop
	})
}

// startWebviewLoop spawns a single OS-thread-locked goroutine that owns all
// WebView2 windows. Win32 windows and COM objects must stay on their creating
// thread, otherwise the message pump deadlocks.
func (u *windowsUI) startWebviewLoop() {
	u.mu.Lock()
	if u.started {
		u.mu.Unlock()
		return
	}
	u.started = true
	u.mu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		for range u.showReq {
			w := webview2.New(false)
			if w == nil {
				// WebView2 runtime not installed, fall back to browser.
				openBrowser(localURL(u.port))
				continue
			}

			u.mu.Lock()
			u.webview = w
			u.mu.Unlock()

			w.SetTitle("Open Todo")
			w.SetSize(800, 600, webview2.HintMin)
			w.Navigate(localURL(u.port))

			// w.Run blocks until the window is destroyed (user clicks X
			// or closeWebview calls w.Destroy from another goroutine).
			w.Run()

			u.mu.Lock()
			u.webview = nil
			u.mu.Unlock()
		}
	}()
}

// showWindow signals the WebView2 loop to create (or re-create) a window.
func (u *windowsUI) showWindow() {
	select {
	case u.showReq <- struct{}{}:
	default:
	}
}

// closeWebview posts WM_CLOSE from the calling goroutine. This is thread-safe
// because Destroy() uses PostMessageW, which can be called from any thread.
func (u *windowsUI) closeWebview() {
	u.mu.Lock()
	w := u.webview
	u.mu.Unlock()
	if w != nil {
		w.Destroy()
	}
}

func (u *windowsUI) Quit() {
	select {
	case u.quitC <- struct{}{}:
	default:
	}
}
