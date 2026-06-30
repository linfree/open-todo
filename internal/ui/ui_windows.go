//go:build windows

package ui

import (
	"github.com/getlantern/systray"
	"github.com/jchv/go-webview2"
)

type windowsUI struct {
	port    int
	webview webview2.WebView
}

func newUI(port int) UI {
	return &windowsUI{port: port}
}

func (u *windowsUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		systray.SetTooltip("Open Todo - 开源待办清单")
		onReady()

		mQuit := systray.AddMenuItem("退出", "退出应用")

		go func() {
			u.webview = webview2.New(false)
			defer u.webview.Destroy()
			u.webview.SetTitle("Open Todo")
			u.webview.SetSize(800, 600, webview2.HintNone)
			u.webview.Navigate(localURL(u.port))
			u.webview.Run()
		}()

		go func() {
			for {
				select {
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {
		if u.webview != nil {
			u.webview.Destroy()
		}
	})
}

func (u *windowsUI) Quit() {
	systray.Quit()
}
