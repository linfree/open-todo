//go:build darwin

package ui

import (
	"github.com/getlantern/systray"
)

type darwinUI struct {
	port int
}

func newUI(port int) UI {
	return &darwinUI{port: port}
}

func (u *darwinUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		onReady()

		mOpen := systray.AddMenuItem("在浏览器中打开", "打开 Web 界面")
		mQuit := systray.AddMenuItem("退出", "退出应用")

		go func() {
			openBrowser(localURL(u.port))
		}()

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					openBrowser(localURL(u.port))
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
}

func (u *darwinUI) Quit() {
	systray.Quit()
}
