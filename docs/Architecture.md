# 架构设计

在 Fyne 里做“代码逻辑不耦合”的核心思路是：**把 Fyne 当成“渲染层 + 事件入口”**，业务逻辑、网络请求、状态管理都放到 UI 外面；UI 只做两件事：

1. 把用户操作转成“意图/命令”（Load / Clear / OpenRepo…）
2. 把状态（repos / loading / error / selection…）渲染出来

Fyne 自身没有强制你用 MVC/MVP/MVVM，但它的 **data binding**（绑定）非常适合走 **MVVM-ish**：View 绑定到 ViewModel 的可观察状态，ViewModel 调用 Service，Service 调用 GitHub API。

---

## 推荐架构：MVVM（或“Presenter/ViewModel + Service”）

### 为什么 MVVM 很适合 Fyne

* Fyne 有 `binding.*`（例如 `binding.String`, `binding.Bool`, `binding.UntypedList`），2.7+ 还可以用 `binding.List[T]` 做类型安全的列表绑定。
* ViewModel 里做异步加载、错误处理、分页策略；View 只订阅变化刷新列表。

> MVP 也行：Presenter 操作 View 接口（`SetLoading/SetError/SetRepos`），但在 Fyne 里你最终还是要回到主线程更新 UI，MVVM 绑定写起来更顺。

---

## github-stars-gui 如何拆分（一个实用的项目结构）

一个清晰、可扩展的拆分：

```
github-stars-gui/
  cmd/stars-gui/main.go          // 启动、组装依赖、显示窗口
  internal/domain/repo.go        // 纯数据结构（Repo / RepoDetails）
  internal/github/client.go      // GitHub API 调用（HTTP、分页、认证）
  internal/app/stars/service.go  // 业务用例（LoadStarred）
  internal/app/repos/service.go  // 业务用例（LoadDetails）
  internal/ui/stars/view.go      // Fyne 视图：输入框、列表、按钮
  internal/ui/stars/vm.go        // ViewModel：状态、命令、并发控制
  internal/ui/details/view.go    // Repo 详情视图
  internal/ui/details/vm.go      // Repo 详情 ViewModel
  internal/ui/nav/navigator.go   // 窗口导航/路由
  internal/ui/widgets/...        // 可复用控件（可选）
```

### 每层职责

* **domain**：纯模型，不引用 Fyne
* **github(client)**：封装 API/HTTP；不引用 Fyne
* **service(usecase)**：组合多个 client、做业务规则；不引用 Fyne
* **ui(view)**：只做布局和绑定，极少 if/for 业务判断
* **ui/vm**：把 service 暴露成 UI 状态+命令；负责并发、取消、错误转换；只引用少量 Fyne binding（可以接受）

这样拆完后，你可以：

* 在不启动 GUI 的情况下单测 `github/client` 和 `service`
* UI 换成 TUI/web 也只需要重写 `ui` 层

---

## 关键最佳实践清单（Fyne 项目里最常踩坑的点）

### 1) UI 更新必须在主线程

在 goroutine 里拉数据没问题，但**更新控件 / binding**最好回主线程：

* 用 `fyne.Do(func(){ ... })` 或 `fyne.DoAndWait(func(){ ... })` 把 UI 更新切回主线程（2.6+）
* 或者只更新 binding（很多情况下 binding 内部是线程安全的，但稳妥起见仍建议用 `fyne.Do` 包裹关键 UI 交互）

### 2) 用 Context 做取消/防抖

用户连续点“加载”时，前一个请求应该取消：

* ViewModel 持有 `cancel()`，新的加载先 cancel 旧的

### 3) UI 不直接依赖 HTTP/JSON

View 不应该 `http.NewRequest`、不应该 `json.Unmarshal`。这些都放 client/service。

### 4) 把状态集中管理（而不是 scattered）

不要到处 `statusLabel.SetText(...)`、`table.Refresh()` scattered。
让 View 绑定到一个统一状态源（ViewModel），状态改变自然驱动刷新。

### 5) 控件拆分：按“可复用/可测试”边界

* 视图拆成：`CredentialsForm`、`RepoList`、`StatusBar`、`Toolbar`
* 子组件通过接口或回调与外界通信，避免直接抓外层变量

---

## 用 github-stars-gui 举一个“MVVM 拆分”的小样

下面不是完整项目，只展示“拆分后长什么样”。

### 1) domain/repo.go（纯模型）

```go
package domain

import "time"

type Repo struct {
	FullName        string
	HTMLURL         string
	Description     string
	Language        string
	Stars           int
	Forks           int
	UpdatedAt       time.Time
	Private         bool
}
```

### 2) github/client.go（API 层接口 + 实现）

```go
package github

import (
	"context"
	"github.com/you/github-stars-gui/internal/domain"
)

type Client interface {
	ListStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error)
}
```

（实现里做 HTTP、分页、header、错误码处理——都留在这一层。）

注意：`/repos/{owner}/{repo}` 需要先把 `fullName` 拆成 `owner` 和 `repo` 再分别 `PathEscape`，不能直接对 `owner/repo` 做 `PathEscape`。

### 3) app/stars/service.go（用例层）

```go
package stars

import (
	"context"
	"github.com/you/github-stars-gui/internal/domain"
	"github.com/you/github-stars-gui/internal/github"
)

type Service struct {
	GH github.Client
}

func (s Service) LoadStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	// 这里可以加缓存、排序、过滤、metrics、重试策略等
	return s.GH.ListStarred(ctx, username, token, perPage)
}
```

### 4) ui/stars/vm.go（ViewModel：状态 + 命令）

这里用 Fyne binding 作为 View 的“数据源”。

```go
package starsui

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/you/github-stars-gui/internal/app/stars"
	"github.com/you/github-stars-gui/internal/domain"
)

type VM struct {
	Username binding.String
	Token    binding.String
	PerPage  binding.String

	Loading binding.Bool
	Status  binding.String
	Error   binding.String // 或者 binding.Untyped 存 error

	Repos binding.UntypedList // 存 []domain.Repo 的每个元素
	// 也可以用 SelectedIndex binding.Int
	selectedIndex int

	svc stars.Service

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewVM(svc stars.Service) *VM {
	vm := &VM{
		Username: binding.NewString(),
		Token:    binding.NewString(),
		PerPage:  binding.NewString(),
		Loading:  binding.NewBool(),
		Status:   binding.NewString(),
		Error:    binding.NewString(),
		Repos:    binding.NewUntypedList(),
		svc:      svc,
	}
	_ = vm.PerPage.Set("100")
	_ = vm.Status.Set("就绪")
	return vm
}

func (vm *VM) Load() {
	vm.mu.Lock()
	if vm.cancel != nil {
		vm.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	vm.cancel = cancel
	vm.mu.Unlock()

	_ = vm.Loading.Set(true)
	_ = vm.Error.Set("")
	_ = vm.Status.Set("加载中…")

	go func() {
		user, _ := vm.Username.Get()
		token, _ := vm.Token.Get()
		perPageStr, _ := vm.PerPage.Get()
		perPage := 100
		// parse perPageStr（略）

		repos, err := vm.svc.LoadStarred(ctx, user, token, perPage)

		// 更新状态（建议回主线程；此处示意，最终在 View 中用 fyne.Do 包裹也行）
		if err != nil {
			_ = vm.Error.Set(err.Error())
			_ = vm.Status.Set("加载失败")
			_ = vm.Loading.Set(false)
			return
		}

		_ = vm.Repos.Set(nil)
		for _, r := range repos {
			_ = vm.Repos.Append(r)
		}
		_ = vm.Status.Set("完成")
		_ = vm.Loading.Set(false)
	}()
}

func (vm *VM) Clear() {
	_ = vm.Repos.Set(nil)
	_ = vm.Status.Set("已清空")
	_ = vm.Error.Set("")
}
```

### 5) ui/stars/view.go（View：布局 + 绑定）

View 只关心：

* 输入框绑定到 `vm.Username/vm.Token/vm.PerPage`
* 按钮触发 `vm.Load/vm.Clear`
* 列表的数据来自 `vm.Repos`

```go
package starsui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewView(w fyne.Window, vm *VM) fyne.CanvasObject {
	username := widget.NewEntryWithData(vm.Username)
	token := widget.NewPasswordEntry()
	token.Bind(vm.Token)

	perPage := widget.NewEntryWithData(vm.PerPage)

	status := widget.NewLabelWithData(vm.Status)
	errLabel := widget.NewLabelWithData(vm.Error)

	loadBtn := widget.NewButton("加载 Stars", func() {
		// 这里可以加 fyne.Do 包裹 vm.Load 的状态更新习惯
		vm.Load()
	})
	clearBtn := widget.NewButton("清空", vm.Clear)

	// 通过 vm.Loading 禁用按钮（需要监听变化）
	vm.Loading.AddListener(binding.NewDataListener(func() {
		loading, _ := vm.Loading.Get()
		if loading {
			loadBtn.Disable()
		} else {
			loadBtn.Enable()
		}
	}))

	list := NewRepoList(vm) // 把 RepoList 抽成子组件

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Username", username),
			widget.NewFormItem("Token", token),
			widget.NewFormItem("per_page", perPage),
		),
		container.NewHBox(loadBtn, clearBtn),
		status,
		errLabel,
	)

	return container.NewBorder(form, nil, nil, nil, list)
}
```

### 6) ui/stars/repo_list.go（可复用组件）

`RepoList` 内部只负责展示 list，并通过回调把“打开 URL”事件抛出去。

如果你想避免 `binding.UntypedList` 的类型断言，可以用 `binding.List[domain.Repo]`（2.7+），或在 VM 内维护一份 `[]domain.Repo` 供点击事件索引。

---

## MVC / MVP / MVVM 怎么选（快速结论）

* **MVC**：在 GUI 里经常“Controller 既管事件又管状态”，容易膨胀；在 Fyne 不如 MVVM 顺手。
* **MVP**：想强单测 Presenter 的时候很香；代价是要写一堆 View 接口方法。
* **MVVM（推荐）**：Fyne binding 让它变得很自然；UI=绑定+布局，VM=状态+命令。

如果你更在意**“UI 完全不引用业务层结构体”**，可以在 VM 层把 `domain.Repo` 转成 `RepoRow`（纯字符串/数字字段），View 只渲染 `RepoRow`。

---

## 一个“组件拆分”的实用规则（很管用）

> **能复用/能单测/有独立状态** 的就拆成组件；其余别拆太碎。

针对 github-stars-gui：

* `CredentialsForm`：username/token/per_page
* `RepoList`：列表展示、选中、双击
* `StatusPanel`：loading/error/status
* `StarsPage`：组合上面三个（页面级容器）
* `AppShell`：菜单、窗口标题、导航（如果后续有多个页面）
