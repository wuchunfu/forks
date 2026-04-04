package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunGit 执行 git 命令（带输出）
func RunGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGitSilent 执行 git 命令（静默，用于 UI 模式），失败时返回 git stderr 内容
func RunGitSilent(args ...string) error {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

// RunGitNoProxy 执行 git 命令，显式禁用代理（用于从本地镜像克隆等场景）
func RunGitNoProxy(args ...string) error {
	// 通过 -c 覆盖全局 git config 中的代理设置
	fullArgs := []string{"-c", "http.proxy=", "-c", "https.proxy="}
	fullArgs = append(fullArgs, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Env = append(os.Environ(),
		"NO_PROXY=localhost,127.0.0.1,0.0.0.0,::1",
		"no_proxy=localhost,127.0.0.1,0.0.0.0,::1",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

// RunGitInDir 在指定目录执行 git 命令（带输出）
func RunGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGitInDirSilent 在指定目录执行 git 命令（静默，用于 UI 模式），失败时返回 git stderr 内容
func RunGitInDirSilent(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

// ResolveAbsDir 将目录转为绝对路径
func ResolveAbsDir(dir string) string {
	if !filepath.IsAbs(dir) {
		absDir, _ := filepath.Abs(dir)
		return absDir
	}
	return dir
}
