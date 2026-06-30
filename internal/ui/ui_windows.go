//go:build windows

package ui

import (
	"github.com/getlantern/systray"
)

type windowsUI struct {
	port int
}

func newUI(port int) UI {
	return &windowsUI{port: port}
}

func (u *windowsUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		systray.SetTooltip("Open Todo - 开源待办清单")
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

func (u *windowsUI) Quit() {
	systray.Quit()
}
