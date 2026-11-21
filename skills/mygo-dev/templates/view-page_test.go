package main

// 视图测试：ui.NewTester 在内存里渲染，不需要窗口，速度和普通单元测试相当。
// 按文本找元素并模拟输入；不要用 sleep 等待动画，改用 tt.Advance。

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCounterPage(t *testing.T) {
	p := &counterPage{}
	tt := ui.NewTester(p.view, 400, 300)

	if err := tt.Click("加一"); err != nil {
		t.Fatal(err)
	}
	if p.n != 1 {
		t.Fatalf("n = %d, want 1", p.n)
	}
	if !tt.HasText("当前值：1") {
		t.Fatalf("texts = %q", tt.Texts())
	}

	if err := tt.Click("清零"); err != nil {
		t.Fatal(err)
	}
	if p.n != 0 || !tt.HasText("当前值：0") {
		t.Fatalf("after reset: n = %d, texts = %q", p.n, tt.Texts())
	}
}
