package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"myshop", true},
		{"my-shop-2", true},
		{"ab", true},
		{"a", false},       // 太短
		{"Myshop", false},  // 大写
		{"1shop", false},   // 数字开头
		{"my_shop", false}, // 下划线不是合法 DNS label
		{"my shop", false}, // 空格
		{"myshop-", false}, // 连字符结尾
		{"base", false},    // 与基底同名
		{"", false},
		{strings.Repeat("a", 41), false},
	}
	for _, c := range cases {
		err := validateName(c.name)
		if c.ok && err != nil {
			t.Errorf("validateName(%q) 应通过，得到 %v", c.name, err)
		}
		if !c.ok && err == nil {
			t.Errorf("validateName(%q) 应报错", c.name)
		}
	}
}

func TestLatestTag(t *testing.T) {
	// v1.10.0 > v1.9.0：字符串比较会搞错，必须按数值
	got, ok := latestTag([]string{"v1.9.0", "v1.10.0", "v1.2.3"})
	if !ok || got != "v1.10.0" {
		t.Fatalf("latestTag = %q, %v; 期望 v1.10.0", got, ok)
	}
	got, ok = latestTag([]string{"v2.0.0", "v10.0.0", "v9.9.9"})
	if !ok || got != "v10.0.0" {
		t.Fatalf("latestTag = %q, %v; 期望 v10.0.0", got, ok)
	}
	// 非正式版标签一律忽略
	if _, ok := latestTag([]string{"v1.0.0-rc1", "tools/create-base/v1.0.0", "nightly"}); ok {
		t.Fatal("不应从非正式版标签中选出版本")
	}
	if _, ok := latestTag(nil); ok {
		t.Fatal("空列表不应返回版本")
	}
}

func TestReplaceOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")

	if err := os.WriteFile(path, []byte("a=1\nb=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replaceOnce(path, "a=1", "a=9"); err != nil {
		t.Fatalf("正常替换失败: %v", err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "a=9\nb=2\n" {
		t.Fatalf("替换结果 = %q", raw)
	}
	// 不存在 -> 报错（模板漂移的信号）
	if err := replaceOnce(path, "missing", "x"); err == nil {
		t.Fatal("占位符不存在时应报错")
	}
	// 出现多次 -> 报错
	if err := os.WriteFile(path, []byte("dup\ndup\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replaceOnce(path, "dup", "x"); err == nil {
		t.Fatal("占位符出现多次时应报错")
	}
}

// fixtureBase 用基底真实的模板文件搭一个最小仓库，并打上标签。
// 用真文件而不是自己编内容，模板里的占位符一旦改名，测试会立刻发现。
func fixtureBase(t *testing.T) string {
	t.Helper()
	if !hasBinary("git") {
		t.Skip("git 不可用")
	}

	// CI runner 通常没有全局 git 身份，而脚手架的 preflight 会（有意地）在缺身份时报错。
	// 用环境变量给整个测试进程一个身份，fixture 仓库和被测的 git 子进程都能继承到。
	t.Setenv("GIT_AUTHOR_NAME", "create-base test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "create-base test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	repoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("不在 git 仓库内")
	}
	root := strings.TrimSpace(string(repoRoot))

	dir := filepath.Join(t.TempDir(), "base")
	files := []string{
		"server/config.yaml.example",
		".deploy.env.example",
		"chuanyun.toml.example",
		"admin/apps/web-antd/.env",
		"Makefile",
	}
	for _, f := range files {
		src := filepath.Join(root, f)
		dst := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := copyFile(src, dst); err != nil {
			t.Fatalf("复制 %s: %v", f, err)
		}
	}

	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"add", "-A"},
		{"commit", "-q", "-m", "base fixture"},
		{"tag", "-a", "v9.9.9", "-m", "v9.9.9"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}
	return strings.TrimSpace(string(out))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestScaffoldFromTag(t *testing.T) {
	base := fixtureBase(t)
	dst := filepath.Join(t.TempDir(), "demo")

	var out bytes.Buffer
	opts := options{
		Name: "demo", Dir: dst, Title: "演示后台",
		BaseURL: base, SkipInstall: true,
	}
	if err := run(opts, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}

	// 远端：base 保留（可以 sync-base），且带 --no-tags 防止污染标签空间
	if got := gitIn(t, dst, "remote", "get-url", "base"); got != base {
		t.Errorf("base remote = %q, 期望 %q", got, base)
	}
	if got := gitIn(t, dst, "config", "remote.base.tagOpt"); got != "--no-tags" {
		t.Errorf("remote.base.tagOpt = %q, 期望 --no-tags", got)
	}
	if err := exec.Command("git", "-C", dst, "remote", "get-url", "origin").Run(); err == nil {
		t.Error("未指定 --origin 时不应有 origin 远端")
	}

	// 标签只在 base/ 命名空间里，没有裸的 v9.9.9
	tags := gitIn(t, dst, "tag", "-l")
	if tags != "base/v9.9.9" {
		t.Errorf("tag -l = %q, 期望仅 base/v9.9.9", tags)
	}

	// 不跟踪基底分支，否则 git push 会推向基底
	if err := exec.Command("git", "-C", dst, "rev-parse", "--abbrev-ref", "main@{upstream}").Run(); err == nil {
		t.Error("main 不应有 upstream")
	}

	// 初始化提交存在，且工作区干净（不该留下未提交的生成物）
	if subject := gitIn(t, dst, "log", "-1", "--format=%s"); subject != "Initialize demo from base v9.9.9" {
		t.Errorf("提交主题 = %q", subject)
	}
	if status := gitIn(t, dst, "status", "--porcelain"); status != "" {
		t.Errorf("工作区应干净，实际:\n%s", status)
	}

	// 后端配置：占位密钥被换成随机值
	cfg := readFile(t, filepath.Join(dst, "server", "config.yaml"))
	if strings.Contains(cfg, "change-this-to-a-strong-secret") {
		t.Error("config.yaml 仍是占位密钥")
	}
	secret := ""
	for _, line := range strings.Split(cfg, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "secret:") {
			secret = strings.Trim(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "secret:")), `"`)
		}
	}
	if len(secret) < 32 {
		t.Errorf("jwt.secret 长度 %d，生产模式要求 >=32", len(secret))
	}

	// 部署配置：项目名三处都换掉
	deployEnv := readFile(t, filepath.Join(dst, ".deploy.env"))
	for _, want := range []string{"PROJECT_NAME=demo", "/opt/demo/server", "/opt/demo/admin"} {
		if !strings.Contains(deployEnv, want) {
			t.Errorf(".deploy.env 缺少 %q", want)
		}
	}
	if strings.Contains(deployEnv, "myproject") {
		t.Error(".deploy.env 仍含 myproject")
	}

	// 前端标识：namespace 必须按项目区分，否则同域名下多个后台互相覆盖登录态
	adminEnv := readFile(t, filepath.Join(dst, "admin", "apps", "web-antd", ".env"))
	for _, want := range []string{"VITE_APP_TITLE=演示后台", "VITE_APP_NAMESPACE=demo-admin"} {
		if !strings.Contains(adminEnv, want) {
			t.Errorf(".env 缺少 %q", want)
		}
	}
	if strings.Contains(adminEnv, "please-replace-me-with-your-own-key") {
		t.Error(".env 仍是占位加密密钥")
	}

	// 默认不生成 chuanyun.toml
	if _, err := os.Stat(filepath.Join(dst, "chuanyun.toml")); !os.IsNotExist(err) {
		t.Error("未指定 --chuanyun 时不应生成 chuanyun.toml")
	}
}

func TestScaffoldFromMainWithOrigin(t *testing.T) {
	base := fixtureBase(t)
	dst := filepath.Join(t.TempDir(), "shop")

	var out bytes.Buffer
	opts := options{
		Name: "shop", Dir: dst, Title: "shop",
		BaseURL: base, Version: "main", Origin: "git@example.com:me/shop.git",
		Chuanyun: true, SkipInstall: true,
	}
	if err := run(opts, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}

	if got := gitIn(t, dst, "remote", "get-url", "origin"); got != "git@example.com:me/shop.git" {
		t.Errorf("origin = %q", got)
	}
	// main 版本的提交信息带短 SHA，便于回溯到具体基底提交
	if subject := gitIn(t, dst, "log", "-1", "--format=%s"); !strings.HasPrefix(subject, "Initialize shop from base main@") {
		t.Errorf("提交主题 = %q", subject)
	}
	chuanyun := readFile(t, filepath.Join(dst, "chuanyun.toml"))
	if !strings.Contains(chuanyun, `project = "shop"`) {
		t.Error("chuanyun.toml 未替换 project")
	}
}

func TestScaffoldRejectsNonEmptyDir(t *testing.T) {
	base := fixtureBase(t)
	dst := filepath.Join(t.TempDir(), "occupied")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(options{Name: "demo", Dir: dst, Title: "demo", BaseURL: base, SkipInstall: true}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("目录非空时应报错")
	}
	// 报错也不能删掉用户已有的文件
	if _, statErr := os.Stat(filepath.Join(dst, "keep.txt")); statErr != nil {
		t.Error("已有文件被误删")
	}
}

func TestScaffoldUnknownVersion(t *testing.T) {
	base := fixtureBase(t)
	dst := filepath.Join(t.TempDir(), "demo")
	err := run(options{Name: "demo", Dir: dst, Title: "demo", BaseURL: base, Version: "v0.0.1", SkipInstall: true}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("指定不存在的版本时应报错")
	}
	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Error("失败后应清理本次创建的目录")
	}
}

func TestParseArgs(t *testing.T) {
	opts, err := parseArgs([]string{"myshop"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if opts.Title != "myshop" {
		t.Errorf("Title 默认应等于项目名，得到 %q", opts.Title)
	}
	if !filepath.IsAbs(opts.Dir) {
		t.Errorf("Dir 应为绝对路径，得到 %q", opts.Dir)
	}
	if opts.BaseURL != defaultBaseURL {
		t.Errorf("BaseURL = %q", opts.BaseURL)
	}

	for _, args := range [][]string{
		{},                           // 缺项目名
		{"a", "b"},                   // 多个项目名
		{"Bad-Name"},                 // 名字不合法
		{"ok", "--version", "1.2.3"}, // 版本号缺 v 前缀
	} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("parseArgs(%v) 应报错", args)
		}
	}
}

// flag 包在第一个位置参数处停止解析，而「项目名在前、选项在后」是最自然的写法，
// 帮助里的示例也是那样写的。两种顺序必须等价。
func TestParseArgsFlagsAfterName(t *testing.T) {
	want := func(t *testing.T, opts options) {
		t.Helper()
		if opts.Name != "myshop" {
			t.Errorf("Name = %q", opts.Name)
		}
		if !opts.SkipInstall {
			t.Error("--skip-install 未生效")
		}
		if opts.Title != "商城后台" {
			t.Errorf("Title = %q", opts.Title)
		}
		if opts.Version != "main" {
			t.Errorf("Version = %q", opts.Version)
		}
	}

	after, err := parseArgs([]string{"myshop", "--skip-install", "--title", "商城后台", "--version", "main"})
	if err != nil {
		t.Fatalf("选项在后: %v", err)
	}
	want(t, after)

	before, err := parseArgs([]string{"--skip-install", "--title", "商城后台", "--version", "main", "myshop"})
	if err != nil {
		t.Fatalf("选项在前: %v", err)
	}
	want(t, before)

	// --key=value 形式不吃掉后面的位置参数
	eq, err := parseArgs([]string{"--title=商城后台", "myshop"})
	if err != nil {
		t.Fatalf("--key=value: %v", err)
	}
	if eq.Name != "myshop" || eq.Title != "商城后台" {
		t.Errorf("--key=value 解析错误: name=%q title=%q", eq.Name, eq.Title)
	}

	// 混在中间也要能认出来
	mid, err := parseArgs([]string{"--origin", "git@example.com:me/x.git", "myshop", "-v"})
	if err != nil {
		t.Fatalf("选项在两侧: %v", err)
	}
	if mid.Name != "myshop" || !mid.Verbose || mid.Origin != "git@example.com:me/x.git" {
		t.Errorf("解析错误: %+v", mid)
	}
}
