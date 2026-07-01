//go:build windows

package ui

import (
	_ "embed"
	"fmt"
	"runtime"
	"sync"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/jchv/go-webview2"
)

//go:embed icon.ico
var trayIcon []byte

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
		systray.SetIcon(trayIcon)

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
		// onExit -- noop
	})
}

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
				openBrowser(localURL(u.port))
				continue
			}

			u.mu.Lock()
			u.webview = w
			u.mu.Unlock()

			w.SetTitle("Open Todo")
			w.SetSize(800, 600, webview2.HintMin)
			w.Navigate(localURL(u.port))

			// Set window icon
			setWindowIconFromWebView(w)

			w.Run()

			u.mu.Lock()
			u.webview = nil
			u.mu.Unlock()
		}
	}()
}

func (u *windowsUI) showWindow() {
	select {
	case u.showReq <- struct{}{}:
	default:
	}
}

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

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	sendMessageW      = user32.NewProc("SendMessageW")
	loadImageW        = user32.NewProc("LoadImageW")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandleW  = kernel32.NewProc("GetModuleHandleW")
)

const (
	WM_SETICON = 0x0080
	ICON_BIG   = 1
	ICON_SMALL = 0
	IMAGE_ICON = 1
)

func setWindowIconFromWebView(w webview2.WebView) {
	hwnd := syscall.Handle(w.Window())
	if hwnd == 0 {
		return
	}

	hInstance, _, _ := getModuleHandleW.Call(0)

	// Try loading icon from exe resources with multiple sizes
	for _, sz := range []uintptr{32, 48, 256, 0} {
		hIcon, _, _ := loadImageW.Call(hInstance, 1, IMAGE_ICON, sz, sz, 0)
		if hIcon != 0 {
			sendMessageW.Call(uintptr(hwnd), WM_SETICON, ICON_BIG, hIcon)
			sendMessageW.Call(uintptr(hwnd), WM_SETICON, ICON_SMALL, hIcon)
			return
		}
	}
}
