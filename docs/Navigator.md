# 窗口管理

在 Fyne 里做“多个界面/页面”，思路和 iOS 的 UIViewController 很像：你需要一个**导航/路由层（Router）**来负责“打开新窗口、切换页面、返回、传参、刷新”。

Fyne 常见的多界面方式有 3 种（按最常用排序）：

1. **同一个 Window 里切换页面**：`container.NewAppTabs(...)` 或 `container.NewStack(...)` 或 `container.NewBorder(...)` 动态替换 center 内容
2. **打开一个新 Window**：`app.NewWindow(...)`（最像 iOS 的 present）
3. **弹窗式详情**：`dialog.NewCustom(...)` / `dialog.NewInformation(...)`（类似 iOS 的 modal sheet）

“点击 repo 弹出详情”最直观的是 **新 Window** 或 **Custom Dialog**。

---

## 推荐做法：引入 Router（路由器）

你想避免耦合的话，不要在 RepoList 里直接 `a.NewWindow(...)`。RepoList 只抛出事件：`OnRepoSelected(repo)`。由 Router 决定“怎么展示详情”。

如果导航可能从 goroutine 触发，记得用 `fyne.Do(...)` 把窗口创建/显示切回主线程。

### Router 接口（UI 层的基础设施）

为避免包的循环引用，`Router` 接口单独放在 `internal/ui/route`；`AppNavigator` 的具体实现仍放在 `internal/ui/nav`。

```go
type Router interface {
	ShowRepoDetails(repoFullName string, token string) // 或者传 RepoID/URL
}
```

ViewModel 不需要知道 Fyne 的具体控件，只要依赖这个接口（或更纯一点：VM 只发事件，View 监听事件再调用 Router）。

主窗口（Stars 列表）也可以由 Router 负责创建，这样 `main.go` 只做依赖组装。

```go
func (n *AppNavigator) ShowStars() {
	w := starsui.NewStarsWindow(n.App, n.StarsSvc, n)
	w.Show()
}
```

应用的主循环可以放在 `main.go` 里调用 `fyneApp.Run()`。

---

## 两种“弹出详情”的实现方式（github-stars-gui 示例）

### 方案 A：打开新窗口（最像 UIViewController push/present）

#### 1) 详情窗口的构造函数（RepoDetailsWindow）

* 接收：`fullName`, `token`
* 内部：有自己的 ViewModel，自己加载 repo 详情，支持刷新

```go
func NewRepoDetailsWindow(a fyne.App, svc RepoService, fullName, token string) fyne.Window {
	w := a.NewWindow("Repo Details: " + fullName)
	w.Resize(fyne.NewSize(900, 600))

	vm := NewRepoDetailsVM(svc, fullName, token, fyne.Do)
	view := NewRepoDetailsView(w, vm)

	w.SetContent(view)
	vm.Load() // 打开即加载
	return w
}
```

#### 2) Stars 列表页点击行 -> Router 打开窗口

```go
type AppNavigator struct {
	App fyne.App
	Svc RepoService
	// 可选：窗口复用缓存，防止重复打开一堆
	details map[string]fyne.Window
}

func (n *AppNavigator) ShowRepoDetails(fullName, token string) {
	if n.details == nil {
		n.details = map[string]fyne.Window{}
	}
	if w, ok := n.details[fullName]; ok {
		w.RequestFocus()
		w.Show()
		return
	}

	w := NewRepoDetailsWindow(n.App, n.Svc, fullName, token)
	n.details[fullName] = w
	w.SetOnClosed(func() { delete(n.details, fullName) })
	w.Show()
}
```

如果 `ShowRepoDetails` 可能并发调用，给 `details` 加锁，或保证只在 UI 主线程调用。

#### 3) 列表的 View 或 VM 触发导航

* RepoList：只调用回调 `onOpen(repo)`
* StarsPage：把回调接到 `nav.ShowRepoDetails(...)`

这样就实现了类似 iOS “打开详情 VC”。

**刷新怎么做？**

* 详情窗口里放一个 “刷新”按钮，调用 `detailsVM.Load()`
* 或者你想“重复打开时自动刷新”：在 `ShowRepoDetails` 里如果窗口已存在，就 `vm.Load()`（你需要能从 window 拿到 vm，或者把 window 包装成 struct 存起来）

---

### 方案 B：Custom Dialog（轻量 modal，适合简单详情）

```go
func ShowRepoDetailsDialog(w fyne.Window, vm *RepoDetailsVM) {
	title := widget.NewLabelWithStyle("Repo Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	refresh := widget.NewButton("刷新", vm.Load)

	content := container.NewBorder(
		container.NewHBox(title, layout.NewSpacer(), refresh),
		nil, nil, nil,
		NewRepoDetailsBody(vm),
	)

	d := dialog.NewCustom("详情", "关闭", content, w)
	d.Resize(fyne.NewSize(800, 500))
	vm.Load()
	d.Show()
}
```

优点：简单、不会打开很多窗口
缺点：交互复杂、需要多页面跳转时不如新窗口清晰

---

## 多页面（像 UINavigationController）怎么做？

Fyne 没有内建 UINavigationController，但你可以用 **stack + history** 自己做。

### 用 Stack 做“push/pop”

```go
type StackNav struct {
	stack *fyne.Container        // container.NewStack()
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

StarsPage push DetailsPage，就像 iOS push VC。

如果你要在详情页返回列表：

* 详情页右上角一个 “返回”按钮 -> `nav.Pop()`

---

## “页面刷新”最佳实践：让每个页面自带 VM，刷新就是 VM.Load()

避免页面之间互相直接改 UI。推荐这条规则：

* **页面（View）只绑定 VM**
* **刷新 = 调用 VM 的命令（Load/Refresh）**
* **传参 = 创建 VM 时注入参数**

比如 RepoDetailsVM：

* Input：`fullName`, `token`
* 可选：`runOnMain`（例如 `fyne.Do`，用于在 goroutine 内更新 UI 绑定）
* Output 状态：`Loading`, `Error`, `Fields(Description, topics, license...)`
* 命令：`Load()`

---

## github-stars-gui：点击 repo 打开详情，你可以这样拆

* `StarsVM`：负责 starred 列表
* `StarsView`：RepoList + 输入框
* `RepoDetailsVM`：负责单 repo 详情（可能调用 `/repos/{owner}/{repo}`、`/repos/.../languages`、`/repos/.../contributors`）
* `RepoDetailsView`：详情页布局 + 刷新按钮
* `Router`：决定是 new window 还是 push/pop

RepoList 的事件链：
`RepoList -> onOpen(repoFullName) -> StarsView -> nav.ShowRepoDetails(repoFullName, token) -> RepoDetailsWindow/Page -> RepoDetailsVM.Load()`

---

## 什么时候选“新窗口” vs “同窗口切换”？

* 详情页信息多、想支持并排对比多个 repo：✅ 新窗口
* 你要做“主界面 + 多子页面 + 返回”：✅ 同窗口 StackNav（更像 app 内导航）
* 详情很简单，只想快速看一眼：✅ dialog
