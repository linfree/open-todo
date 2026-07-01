//go:build windows

package ui

import (
	_ "embed"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

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

			// Set window icon in background (wait for window creation)
			go setWindowIcon()

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
	findWindowW       = user32.NewProc("FindWindowW")
	sendMessageW      = user32.NewProc("SendMessageW")
	loadImageW        = user32.NewProc("LoadImageW")
)

const (
	WM_SETICON = 0x0080
	ICON_BIG   = 1
	ICON_SMALL = 0
	IMAGE_ICON = 1
	LR_LOADFROMFILE = 0x0010
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	getModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
)

func setWindowIcon() {
	// Wait for window creation
	time.Sleep(500 * time.Millisecond)

	// Get HINSTANCE of current exe (pass nil for current process)
	hInstance, _, _ := getModuleHandleW.Call(0)

	for i := 0; i < 20; i++ {
		title, _ := syscall.UTF16PtrFromString("Open Todo")
		hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(title)))
		if hwnd != 0 {
			// Load icon from exe resource (ID=1, embedded by rsrc)
			hIcon, _, _ := loadImageW.Call(hInstance, 1, IMAGE_ICON, 32, 32, 0)
			if hIcon == 0 {
				hIcon, _, _ = loadImageW.Call(hInstance, 1, IMAGE_ICON, 16, 16, 0)
			}

			// Fallback: try loading from file
			if hIcon == 0 {
				iconPath := syscall.StringToUTF16Ptr("internal\\ui\\icon.ico")
				hIcon, _, _ = loadImageW.Call(0, uintptr(unsafe.Pointer(iconPath)), IMAGE_ICON, 0, 0, LR_LOADFROMFILE)
			}

			if hIcon != 0 {
				sendMessageW.Call(hwnd, WM_SETICON, ICON_BIG, hIcon)
				sendMessageW.Call(hwnd, WM_SETICON, ICON_SMALL, hIcon)
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
