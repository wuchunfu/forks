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

// RunGitInDir 在指定目录执行 git 命令（带输出）
func RunGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGitInDirSilent 在指定目录执行 git 命令（静默），失败时返回 git stderr 内容
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
