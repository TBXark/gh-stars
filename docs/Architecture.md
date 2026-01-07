# 架构设计

目标：在 Fyne 中做到"UI 轻量、逻辑可测、层次清晰"。核心做法是把 Fyne 当成**渲染层 + 事件入口**，业务逻辑、网络请求、状态管理全部放在 UI 外面；UI 只负责把用户意图转成命令、把状态渲染出来。

本项目推荐 **MVVM-ish**（也可叫 Presenter + Service）：View 绑定 ViewModel 的可观察状态，ViewModel 调用 Service，Service 调用 GitHub API。

---

## 设计目标与约束

- UI 不直接依赖 HTTP/JSON，避免耦合和难测
- 业务逻辑可单测（不启动 GUI 也能测）
- 支持取消、重试、分页等行为，且状态可观察
- Fyne 主线程安全：UI 更新必须回主线程

---

## 架构总览

依赖方向（只能单向）：

```
ui(view) -> ui(vm) -> app(service) -> github(client) -> domain
```

层级职责：

- domain：纯模型，仅数据结构
- github(client)：API/HTTP 细节（分页、鉴权、错误码）
- app(service)：业务用例（组合 client、策略、缓存/重试）
- ui(vm)：状态 + 命令 + 并发/取消
- ui(view)：布局 + 绑定 + 事件入口

---

## 建议的项目结构

```
cmd/stars-gui/main.go          // 启动、组装依赖、显示窗口
internal/domain/repo.go        // 纯数据结构（Repo / RepoDetails）
internal/github/client.go      // GitHub API 调用（HTTP、分页、认证）
internal/app/stars/service.go  // 业务用例（LoadStarred）
internal/app/repos/service.go  // 业务用例（LoadDetails）
internal/ui/stars/view.go      // Fyne 视图：输入框、列表、按钮
internal/ui/stars/vm.go        // ViewModel：状态、命令、并发控制
internal/ui/details/view.go    // Repo 详情视图
internal/ui/details/vm.go      // Repo 详情 ViewModel
internal/ui/route/router.go    // Router 接口（避免循环依赖）
internal/ui/nav/navigator.go   // 窗口导航/路由实现
internal/ui/widgets/...        // 可复用控件（可选）
```

---

## 关键交互流程（以"加载 Stars"为例）

1) View 触发 `vm.Load()`
2) VM 取消旧请求，创建新 `context`，更新 `Loading/Status/Error`
3) VM 调用 `service.LoadStarred()`
4) service 调用 `github.Client.ListStarred()`
5) VM 接收结果，更新绑定状态
6) View 通过 binding 自动刷新列表

---

## ViewModel 的职责边界

- 聚合 UI 状态：loading/status/error/selection/list
- 处理并发与取消：新请求先 cancel 旧请求
- 转换错误为可展示信息
- 负责分页、重试、过滤等 UI 相关策略

ViewModel 可以引用少量 Fyne binding，但不应依赖具体控件。

---

## UI 线程与并发

- 在 goroutine 中拉数据没有问题
- **更新控件或 binding 时必须回主线程**

推荐做法：

- `fyne.Do(func(){ ... })` 或 `fyne.DoAndWait(func(){ ... })`
- 若只是更新 binding，尽量在 `fyne.Do` 中集中完成

---

## 数据绑定策略

- 2.7+ 建议用 `binding.List[T]` 做类型安全列表
- 低版本可用 `binding.UntypedList`，但注意类型断言
- 如果希望 UI 不依赖 `domain.Repo`，可在 VM 层转换成 `RepoRow`

---

## 最小可行示例（MVVM 拆分）

### domain/repo.go

```go
package domain

import "time"

type Repo struct {
	FullName    string
	HTMLURL     string
	Description string
	Language    string
	Stars       int
	Forks       int
	UpdatedAt   time.Time
	Private     bool
}
```

### github/client.go

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

注意：`/repos/{owner}/{repo}` 需要先把 `fullName` 拆成 `owner` 和 `repo` 再分别 `PathEscape`，不能直接对 `owner/repo` 做 `PathEscape`。

### app/stars/service.go

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
	return s.GH.ListStarred(ctx, username, token, perPage)
}
```

### ui/stars/vm.go

```go
package starsui

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2/data/binding"
	"github.com/you/github-stars-gui/internal/app/stars"
)

type VM struct {
	Username binding.String
	Token    binding.String
	PerPage  binding.String

	Loading binding.Bool
	Status  binding.String
	Error   binding.String

	Repos binding.UntypedList

	svc    stars.Service
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
	_ = vm.Status.Set("加载中...")

	go func() {
		user, _ := vm.Username.Get()
		token, _ := vm.Token.Get()
		perPageStr, _ := vm.PerPage.Get()
		_ = perPageStr // parse perPageStr (略)

		repos, err := vm.svc.LoadStarred(ctx, user, token, 100)

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

### ui/stars/view.go

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
		vm.Load()
	})
	clearBtn := widget.NewButton("清空", vm.Clear)

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

	return container.NewBorder(form, nil, nil, nil, NewRepoList(vm))
}
```

---

## 组件拆分规则（实用）

> 能复用/能单测/有独立状态 的就拆成组件；其余不要拆太碎。

建议组件：

- CredentialsForm：username/token/per_page
- RepoList：列表展示、选中、双击
- StatusPanel：loading/error/status
- StarsPage：组合上面三个
- AppShell：菜单、窗口标题、导航（如果有多页面）

---

## MVC / MVP / MVVM 的快速结论

- MVC：Controller 容易膨胀，在 Fyne 里不太顺
- MVP：单测 Presenter 很好，但要写一堆 View 接口
- MVVM（推荐）：Fyne binding 天然适配，View 更轻
