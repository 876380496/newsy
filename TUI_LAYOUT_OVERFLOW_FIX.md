# TUI 布局溢出问题：lipgloss 分层渲染模式

## 问题描述

在基于 Bubble Tea + lipgloss 的 TUI 应用中，多窗格布局出现上半部分被隐藏、底部边框消失的问题。根本原因是内容超长时撑破了窗格的高度限制。

## 根因分析

### 1. `Height()` 不裁剪内容

lipgloss 的 `Height()` 方法只在内容不足时填充空白行，**内容超长时直接穿透不裁剪**。

源码证据（`style.go:425-427` → `align.go:62-66`）：

```go
func alignTextVertical(str string, pos Position, height int, _ *termenv.Style) string {
    strHeight := strings.Count(str, "\n") + 1
    if height < strHeight {
        return str  // 内容比设定高度高 → 直接返回，不裁剪！
    }
    // ... 填充逻辑仅对短内容生效
}
```

真正执行裁剪的是 `MaxHeight()`（`style.go:461-467`）。

### 2. `MaxHeight` 在 `border` 之后执行

lipgloss `Render()` 内部的执行顺序：

```
1. Word wrap（Width 约束）
2. 文本样式渲染
3. Padding（上下左右）
4. Height() 填充/对齐      ← 不裁剪！
5. 水平对齐
6. applyBorder()            ← 边框在此添加
7. applyMargins()
8. MaxWidth() 裁剪
9. MaxHeight() 裁剪         ← 裁剪在最后，边框会被裁掉！
```

**如果在同一个带边框的 style 上设置 `MaxHeight`，裁剪会把边框一起裁掉。**

### 3. 初始渲染时机

Bubble Tea 启动时，`WindowSizeMsg` 到达之前：
- `m.width = 0, m.height = 0`
- `resize()` 提前返回，list 保持在 0×0 尺寸
- `bubbles/list` 的 `View()` 在尺寸为 0 时可能渲染全部条目
- 进一步加剧溢出

## 解决方案：内层裁剪 + 外层边框

参考 [bubbles viewport PR #228](https://github.com/charmbracelet/bubbles/pull/228/files) 的规范模式。

### 核心原则

> **尺寸约束（Width/Height/MaxHeight）放在无边框的内层 style 上，边框 style 只负责装饰，不参与尺寸控制。**

### 正确模式

```go
// 1. 计算内容区可用尺寸（扣除边框和 padding）
innerHeight := totalHeight - paneStyle.GetVerticalFrameSize()
innerWidth  := totalWidth  - paneStyle.GetHorizontalFrameSize()

// 2. 内层 style：无边框，只负责尺寸约束和裁剪
innerContent := lipgloss.NewStyle().
    Width(innerWidth).
    Height(innerHeight).      // 短内容填充
    MaxHeight(innerHeight).   // 长内容裁剪（在边框添加前执行）
    Render(rawContent)

// 3. 外层 style：仅有边框和 padding，包裹已裁剪好的内容
finalView := paneStyle.Render(innerContent)
```

### 错误模式（会导致问题）

```go
// 尺寸约束和边框混在同一个 style 上
finalView := paneStyle.
    Width(innerWidth).
    Height(innerHeight).
    MaxHeight(innerHeight).   // 边框先加上，然后 MaxHeight 裁剪 → 边框被裁掉！
    Render(rawContent)
```

## 本项目修复位置

`internal/ui/view.go:57-76` — 三个窗格均采用分层渲染：

- `sourceContent` / `articleContent` / `previewContent`：内层无边框 style，负责 Width/Height/MaxHeight
- `sourcePane` / `articlePane` / `previewPaneStyle`：外层边框 style，仅 `Render(innerContent)`

## 参考

- [bubbles viewport PR #228 — fix(viewport): properly truncate to size](https://github.com/charmbracelet/bubbles/pull/228/files)
- [lipgloss 源码 `style.go` Render 方法](https://github.com/charmbracelet/lipgloss/blob/master/style.go)
- [bubbles list 组件 FAQ](https://github.com/charmbracelet/bubbles/pull/359)
