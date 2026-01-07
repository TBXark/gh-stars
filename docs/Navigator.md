# 窗口管理

目标：让窗口/页面的创建、切换、复用、关闭集中在一个导航层，避免 View/VM 直接操纵窗口，降低耦合并提升可测试性。

---

## 一、窗口管理方案（概览）

Fyne 常见的多界面方式：

1. 同一个 Window 内切换页面（`container.NewStack` / `container.NewAppTabs`）
2. 打开新 Window（`app.NewWindow`）
3. 弹窗（`dialog.NewCustom` / `dialog.NewInformation`）

"点击 repo 打开详情"通常用新窗口或自定义 dialog。

### 1) 新窗口（独立任务）

- 适合详情信息多、交互复杂、可长期停留的页面
- 更像 iOS 的 present/push，用户心智清晰
- 建议配合 Router 做窗口复用，避免重复打开

### 2) 弹窗（轻量确认/查看）

- 适合信息量小、无需跳转的页面
- 优点是轻量、不打断主窗口结构
- 缺点是交互复杂时不够友好

### 3) 同窗口多页面

适合页面数量固定、需要快速切换的场景。常见实现：Stack 和 AppTabs。

#### Stack 方案（类似 UINavigationController）

如果需要 push/pop 页面：

```go
type StackNav struct {
	stack *fyne.Container
	hist  []fyne.CanvasObject
}

func NewStackNav(root fyne.CanvasObject) *StackNav {
	s := container.NewStack(root)
	return &StackNav{stack: s, hist: []fyne.CanvasObject{root}}
}

func (n *StackNav) Push(page fyne.CanvasObject) {
	n.hist = append(n.hist, page)
	n.stack.Objects = []fyne.CanvasObject{page}
	n.stack.Refresh()
}

func (n *StackNav) Pop() {
	if len(n.hist) <= 1 {
		return
	}
	n.hist = n.hist[:len(n.hist)-1]
	n.stack.Objects = []fyne.CanvasObject{n.hist[len(n.hist)-1]}
	n.stack.Refresh()
}
```

#### AppTabs 方案（顶部 Tab）

当页面数量固定、需要快速切换时，`container.NewAppTabs` 最省心。

适用场景：

- Stars 列表、收藏夹、设置这类固定页面
- 不需要复杂返回栈的页面结构

简单示例：

```go
func NewMainTabs(router Router, stars fyne.CanvasObject, settings fyne.CanvasObject) *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem("Stars", stars),
		container.NewTabItem("Settings", settings),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	return tabs
}
```

如果需要在 tab 内打开详情：

- 方案 A：tab 内用 `container.NewStack` 做 push/pop（详情页覆盖列表）
- 方案 B：点击详情仍开新窗口（详情是独立任务）
- 方案 C：详情页作为一个临时 Tab（用完删除）

---

## 二、导航层设计（Router + Navigator）

### 设计原则

- 任何 window 的创建/关闭/复用，都由 Router 处理
- View/VM 只发事件，不直接调用 `app.NewWindow`
- 所有 UI 创建与更新都在主线程执行（`fyne.Do`）
- 详情窗口支持复用，避免重复打开同一 repo

### Router 接口与位置

把 `Router` 接口放在 `internal/ui/route`，具体实现放在 `internal/ui/nav`，避免循环依赖。

```go
type Router interface {
	ShowStars()
	ShowRepoDetails(fullName, token string)
}
```

`main.go` 只负责组装依赖，最终调用 `nav.ShowStars()` 和 `app.Run()`。

### 窗口生命周期和复用策略

建议：

- Stars 主窗口：单例，只创建一次
- Repo 详情窗口：按 fullName 复用，打开已存在窗口则聚焦
- 关闭窗口时从缓存移除，避免泄漏

示例：

```go
type AppNavigator struct {
	App         fyne.App
	StarsSvc    StarsService
	RepoSvc     RepoService
	starsWindow fyne.Window
	details     map[string]*RepoDetailsWindow
}

func (n *AppNavigator) ShowStars() {
	fyne.Do(func() {
		if n.starsWindow == nil {
			w := NewStarsWindow(n.App, n.StarsSvc, n)
			n.starsWindow = w
		}
		n.starsWindow.Show()
		n.starsWindow.RequestFocus()
	})
}

func (n *AppNavigator) ShowRepoDetails(fullName, token string) {
	fyne.Do(func() {
		if n.details == nil {
			n.details = map[string]*RepoDetailsWindow{}
		}
		if w, ok := n.details[fullName]; ok {
			w.Window.Show()
			w.Window.RequestFocus()
			w.Refresh() // 可选: 重新加载
			return
		}

		w := NewRepoDetailsWindow(n.App, n.RepoSvc, fullName, token)
		n.details[fullName] = w
		w.Window.SetOnClosed(func() { delete(n.details, fullName) })
		w.Window.Show()
	})
}
```

`RepoDetailsWindow` 是一个轻量 wrapper，避免把 VM 藏在 `fyne.Window` 里。

```go
type RepoDetailsWindow struct {
	Window fyne.Window
	VM     *RepoDetailsVM
}

func NewRepoDetailsWindow(a fyne.App, svc RepoService, fullName, token string) *RepoDetailsWindow {
	w := a.NewWindow("Repo Details: " + fullName)
	w.Resize(fyne.NewSize(900, 600))

	vm := NewRepoDetailsVM(svc, fullName, token, fyne.Do)
	view := NewRepoDetailsView(w, vm)

	w.SetContent(view)
	vm.Load()
	return &RepoDetailsWindow{Window: w, VM: vm}
}

func (w *RepoDetailsWindow) Refresh() {
	w.VM.Load()
}
```

### 事件流示例

- `RepoList` 仅触发 `onOpen(repo)`
- `StarsView` 将事件转发给 `Router.ShowRepoDetails()`
- Router 负责窗口复用、聚焦、刷新

这样 RepoList 和 ViewModel 不需要关心窗口管理细节。

---

## 三、页面刷新最佳实践

- 页面只绑定 VM
- 刷新 = 调用 VM 的命令（`Load/Refresh`）
- 传参 = 创建 VM 时注入参数

示例输入/输出：

- Input: `fullName`, `token`, `runOnMain`
- State: `Loading`, `Error`, `Fields`
- Command: `Load()`

---

## 四、github-stars-gui 的推荐拆分

- StarsVM：负责 starred 列表
- StarsView：RepoList + 输入框
- RepoDetailsVM：负责详情与刷新
- RepoDetailsView：布局 + 刷新按钮
- Router/Navigator：窗口复用与跳转
