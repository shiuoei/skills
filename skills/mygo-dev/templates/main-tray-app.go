package main

// 托盘常驻应用骨架（原生 UI）。演示：
//   - 单实例：第二个实例启动时唤起已有窗口；
//   - 关闭窗口只隐藏，应用继续在托盘运行；
//   - 托盘菜单：打开窗口 / 检查更新 / 退出；
//   - 图标按用途缩放：托盘 32×32（PNG 宽度即 HICON 尺寸），窗口 256×256。
//
// 使用前替换 resources/icon.png 为你的 1024×1024 方形图标；
// 自动更新接入见同目录 updater.go 与 references/12-distribution-updates-ci.md。

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/icon.png
var appIconPNG []byte

type appState struct {
	window   *mygo.Window
	tray     *mygo.Tray
	quitting bool
}

func (a *appState) view(c *ui.Context) {
	ui.Column(c).Fill().Center().Gap(12).Children(func() {
		ui.Text(c, "托盘应用").FontSize(22).Bold()
		ui.Text(c, "关闭这个窗口不会退出应用，它会继续在托盘运行。")
		if ui.PrimaryButton(c, "隐藏窗口").Clicked() {
			a.hide()
		}
	})
}

func (a *appState) show() {
	if a.window == nil || a.window.IsDestroyed() {
		return
	}
	a.window.Show()
	a.window.Focus()
}

func (a *appState) hide() {
	if a.window != nil && !a.window.IsDestroyed() {
		a.window.Hide()
	}
}

func main() {
	// 同一时间只允许一个实例；拿不到锁说明已有实例在运行。
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}

	state := &appState{}
	mygo.App.OnSecondInstance(func(_ []string, _ string) { state.show() })

	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Your App",
			Width:   900,
			Height:  640,
			Content: ui.View(state.view),
		})
		state.window = w
		if windowIcon, err := iconAt(appIconPNG, 256); err == nil {
			_ = w.SetIcon(windowIcon)
		}
		w.OnClose(func(e *mygo.CloseEvent) {
			if state.quitting {
				return
			}
			e.PreventDefault()
			w.Hide()
		})
		w.Show()

		// 托盘图标的 PNG 宽度就是 HICON 尺寸：32×32 像素即可。
		trayIcon, err := iconAt(appIconPNG, 32)
		if err != nil {
			log.Fatalf("resize tray icon: %v", err)
		}
		tray, err := mygo.NewTray(mygo.TrayOptions{
			ID:      "main",
			Icon:    trayIcon,
			ToolTip: "Your App",
			Menu: mygo.NewMenu([]*mygo.MenuItem{
				{Label: "打开窗口", Click: func(*mygo.MenuItem, *mygo.Window) { state.show() }},
				mygo.Separator(),
				{Label: "退出", Click: func(*mygo.MenuItem, *mygo.Window) {
					state.quitting = true
					mygo.App.Quit()
				}},
			}),
		})
		if err != nil {
			// 托盘失败时应用仍能运行，但关闭窗口后无法再打开：必须记录日志。
			log.Printf("create tray: %v", err)
			return
		}
		state.tray = tray
	})

	// 退出前清理：关闭数据库、停止后台任务、保存状态。
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) {})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// iconAt 把 PNG 缩放为 size×size：按面积平均，并按覆盖率加权颜色，
// 避免透明边缘缩小后出现黑边。
func iconAt(src []byte, size int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("decode icon: %w", err)
	}
	b := img.Bounds()
	if b.Dx() == size && b.Dy() == size {
		return src, nil
	}
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		y0, y1 := b.Min.Y+y*b.Dy()/size, b.Min.Y+(y+1)*b.Dy()/size
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < size; x++ {
			x0, x1 := b.Min.X+x*b.Dx()/size, b.Min.X+(x+1)*b.Dx()/size
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, bl, a uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					c := color.NRGBAModel.Convert(img.At(sx, sy)).(color.NRGBA)
					r += uint64(c.R) * uint64(c.A)
					g += uint64(c.G) * uint64(c.A)
					bl += uint64(c.B) * uint64(c.A)
					a += uint64(c.A)
				}
			}
			if a == 0 {
				continue
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r/a + 1/2),
				G: uint8(g/a + 1/2),
				B: uint8(bl/a + 1/2),
				A: uint8(a/uint64((x1-x0)*(y1-y0)) + 1/2),
			})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, fmt.Errorf("encode icon: %w", err)
	}
	return out.Bytes(), nil
}
