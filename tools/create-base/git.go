package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// runner 在固定目录下执行命令。verbose 时把命令本身打出来，方便复现。
type runner struct {
	dir     string
	out     io.Writer
	verbose bool
}

// git 执行一条 git 命令并返回去掉首尾空白的 stdout。
func (r runner) git(args ...string) (string, error) {
	return r.capture("git", args...)
}

// capture 执行命令并返回 stdout；失败时把 stderr 带进 error，否则排查全靠猜。
func (r runner) capture(name string, args ...string) (string, error) {
	if r.verbose {
		fmt.Fprintf(r.out, "    $ %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = r.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("%s %s: %w%s", name, strings.Join(args, " "), err, indentBlock(msg))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// stream 执行命令并把输出直接透传（依赖安装这类耗时长、需要看进度的命令用它）。
func (r runner) stream(name string, args ...string) error {
	if r.verbose {
		fmt.Fprintf(r.out, "    $ %s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = r.dir
	cmd.Stdout = r.out
	cmd.Stderr = r.out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func indentBlock(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "      " + l
	}
	return "\n" + strings.Join(lines, "\n")
}

func hasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
