package internal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 克隆阶段
type clonePhase int

const (
	phaseResolving clonePhase = iota
	phasePreparing
	phaseCloning
	phaseRemotes
	phaseDone
	phaseError
)

// lipgloss 样式
var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	styleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleLabel   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	styleURL     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleHint    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// cloneModel bubbletea 状态机
type cloneModel struct {
	phase      clonePhase
	spinner    spinner.Model
	info       *RepoInfo
	targetDir  string
	useMirror  bool
	prepareMsg string
	err        error
}

type phaseMsg struct {
	phase clonePhase
}

type prepareDoneMsg struct {
	result PrepareResult
}

type cloneDoneMsg struct {
	useMirror bool
	err       error
}

type remotesDoneMsg struct {
	err error
}

func newCloneModel(info *RepoInfo, targetDir string) cloneModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	return cloneModel{
		phase:     phaseResolving,
		spinner:   s,
		info:      info,
		targetDir: targetDir,
	}
}

func (m cloneModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return phaseMsg{phase: phasePreparing}
	})
}

func (m cloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case phaseMsg:
		m.phase = msg.phase
		if m.phase == phaseDone || m.phase == phaseError {
			return m, tea.Quit
		}
		return m, nil

	case prepareDoneMsg:
		m.prepareMsg = msg.result.Message
		m.useMirror = msg.result.UseMirror
		m.phase = phaseCloning
		return m, nil

	case cloneDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.phase = phaseError
			return m, tea.Quit
		}
		m.useMirror = msg.useMirror
		m.phase = phaseRemotes
		return m, nil

	case remotesDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.phase = phaseDone
		return m, tea.Quit
	}

	return m, nil
}

func (m cloneModel) View() string {
	switch m.phase {
	case phaseResolving:
		return fmt.Sprintf("  %s 正在解析仓库地址...", m.spinner.View())

	case phasePreparing:
		serverDisplay := m.info.ServerOrigin
		if serverDisplay == "" {
			serverDisplay = "(未配置)"
		}
		return fmt.Sprintf("  %s 正在联系镜像服务 %s ...", m.spinner.View(), serverDisplay)

	case phaseCloning:
		if m.useMirror {
			return fmt.Sprintf("  %s 正在克隆仓库（镜像加速）...", m.spinner.View())
		}
		return fmt.Sprintf("  %s 正在克隆仓库（直连）...", m.spinner.View())

	case phaseRemotes:
		return fmt.Sprintf("  %s 正在设置 remote...", m.spinner.View())

	case phaseDone, phaseError:
		// 最终结果在 alt screen 恢复后由 printResult 打印
		return fmt.Sprintf("  %s 完成", m.spinner.View())
	}

	return ""
}

// RunCloneUI 启动 bubbletea 克隆流程（带动画）
func RunCloneUI(info *RepoInfo, targetDir, token string, force bool) error {
	model := newCloneModel(info, targetDir)

	// 收集结果
	type stepResult struct {
		prepareResult PrepareResult
		cloneErr      error
		remotesErr    error
	}
	result := &stepResult{}

	done := make(chan struct{})

	p := tea.NewProgram(model)

	// 在 goroutine 中逐步发送消息驱动状态机
	go func() {
		// 1. 准备阶段 - 调用镜像服务 prepare API
		var prepareResult PrepareResult
		if info.ServerOrigin != "" {
			prepareResult = PrepareFromServer(info.ServerOrigin, token, info.Source, info.Author, info.Repo, force)
		}
		result.prepareResult = prepareResult
		p.Send(prepareDoneMsg{result: prepareResult})

		// 2. 克隆阶段 - 静默执行 git clone（alt screen 会隐藏 git 输出）
		useMirror := prepareResult.UseMirror
		cloneURL := info.OriginalURL
		if useMirror {
			cloneURL = info.MirrorURL
		}
		var cloneErr error
		if useMirror {
			// 镜像克隆是本地连接，需要禁用代理避免干扰
			cloneErr = RunGitNoProxy("clone", cloneURL, targetDir)
		} else {
			cloneErr = RunGitSilent("clone", cloneURL, targetDir)
		}
		if cloneErr != nil {
			result.cloneErr = cloneErr
			p.Send(cloneDoneMsg{useMirror: useMirror, err: cloneErr})
			close(done)
			return
		}
		p.Send(cloneDoneMsg{useMirror: useMirror, err: nil})

		// 3. 设置 remotes
		cloneDir := ResolveAbsDir(targetDir)
		var remotesErr error

		if err := RunGitInDirSilent(cloneDir, "remote", "set-url", "origin", info.OriginalURL); err != nil {
			remotesErr = fmt.Errorf("设置 remote origin 失败: %w", err)
		}

		result.remotesErr = remotesErr
		p.Send(remotesDoneMsg{err: remotesErr})
		close(done)
	}()

	_, err := p.Run()
	<-done

	// alt screen 已恢复，在普通终端打印最终结果
	if result.cloneErr != nil {
		printError(info, result.prepareResult.Message, result.cloneErr)
		return result.cloneErr
	}

	printDone(info)
	return err
}

// printDone 在普通终端打印成功结果
func printDone(info *RepoInfo) {
	fmt.Println()
	fmt.Println(styleSuccess.Render("  ✓ 克隆完成！"))
	fmt.Println()
	fmt.Printf("  %s  →  %s\n", styleLabel.Render("origin"), styleURL.Render(info.OriginalURL))
	fmt.Println()
}

// printError 在普通终端打印错误结果
func printError(info *RepoInfo, prepareMsg string, cloneErr error) {
	fmt.Println()
	fmt.Println(styleError.Render("  ✗ 克隆失败"))
	fmt.Println()

	if prepareMsg != "" {
		fmt.Println(styleWarning.Render(fmt.Sprintf("  %s", prepareMsg)))
		fmt.Println()
	}

	fmt.Println(styleDim.Render(fmt.Sprintf("  %v", cloneErr)))
	fmt.Println()

	if info.ServerOrigin == "" {
		fmt.Println(styleHint.Render("  提示: fclone config server <url>"))
		fmt.Println()
	}
}
