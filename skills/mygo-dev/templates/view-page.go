package main

// 原生 UI 页面的写法：状态放在应用结构体里，view 每帧按状态重建界面。
// 复合页面可以像 counterPage 这样拆成独立的类型和方法。

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

type counterPage struct {
	n int
}

func (p *counterPage) view(c *ui.Context) {
	ui.Column(c).Fill().Gap(12).Children(func() {
		ui.Text(c, "计数器").FontSize(20).Bold()
		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, "加一").Clicked() {
				p.n++
			}
			if ui.Button(c, "清零").Clicked() {
				p.n = 0
			}
		})
		ui.Text(c, fmt.Sprintf("当前值：%d", p.n))
	})
}
