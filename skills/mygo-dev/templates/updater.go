package main

// 签名自动更新接入（原生 UI 应用用 plugins/updater/native，不需要 webview）。
//
// 前提：mygo.json 的 updates 配置了 publicKey 与 github/url；构建时提供私钥
// （MYGO_UPDATER_PRIVATE_KEY 或 updates.privateKey）；发布说明文件有 `## <version>`。
// 自动更新只对打包构建生效，安装目录必须可写。

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/egoist/mygo/ui"
)

// registerUpdater 注册原生更新窗口。必须在 mygo.App.Run 之前调用一次。
// language 是窗口语言（如 "zh-CN"），插件注册后固定；切换语言需重启应用。
func registerUpdater(language string, iconPNG []byte) {
	mygo.Use(native.New(updater.Options{
		Icon:     iconPNG,
		Language: language,
	}))
}

// updaterMenuItem 供菜单使用；不要重复实现"检查更新"的逻辑。
func updaterMenuItem() *mygo.MenuItem {
	return updater.MenuItem()
}

// updatePanel 是"关于"页里的更新面板：显示状态、自动检查开关、手动检查。
type updatePanel struct {
	automatic bool
	available string
	lastCheck time.Time
}

// refresh 读取插件状态；可在每帧调用。
func (p *updatePanel) refresh() {
	p.automatic = updater.AutomaticChecks()
	p.available = updater.AvailableVersion()
	p.lastCheck = updater.LastCheck()
}

func (p *updatePanel) view(c *ui.Context) {
	p.refresh()
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, "软件更新").FontSize(20).Bold()
		switch {
		case p.available != "":
			ui.Text(c, "可用版本："+p.available)
		case p.lastCheck.IsZero():
			ui.Text(c, "尚未检查更新")
		default:
			ui.Text(c, "上次检查："+p.lastCheck.Format("2006-01-02 15:04"))
		}
		if ui.Checkbox(c, &p.automatic, "启动时自动检查").Changed() {
			updater.SetAutomaticChecks(p.automatic)
		}
		if ui.Button(c, "检查更新…").Clicked() {
			updater.CheckForUpdates()
		}
	})
}

// watchUpdater 在用户改了更新偏好时回调，例如刷新"关于"页。
func watchUpdater(onChange func()) (off func()) {
	return updater.OnChange(onChange)
}
