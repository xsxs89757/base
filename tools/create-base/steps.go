package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// preflight 把所有「晚失败会很难受」的前置条件提前检查掉：
// 尤其是 git 身份，缺了会等到最后一步 commit 才报错，那时已经装完依赖了。
func preflight(s *session) error {
	if !hasBinary("git") {
		return fmt.Errorf("未找到 git")
	}
	probe := runner{dir: ".", out: s.out, verbose: s.opts.Verbose}
	if ident, err := probe.git("var", "GIT_COMMITTER_IDENT"); err != nil || strings.TrimSpace(ident) == "" {
		return fmt.Errorf("git 身份未配置，请先执行:\n      git config --global user.name  \"你的名字\"\n      git config --global user.email \"你的邮箱\"")
	}
	if !s.opts.SkipInstall {
		if !hasBinary("go") {
			return fmt.Errorf("未找到 go（或用 --skip-install 跳过依赖安装）")
		}
		if !hasBinary("pnpm") {
			return fmt.Errorf("未找到 pnpm（或用 --skip-install 跳过依赖安装）")
		}
	}

	info, err := os.Stat(s.opts.Dir)
	switch {
	case os.IsNotExist(err):
		// 由 git clone 创建，记下来以便失败时清理
		s.createdDir = true
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s 已存在且不是目录", s.opts.Dir)
	default:
		empty, err := isDirEmpty(s.opts.Dir)
		if err != nil {
			return err
		}
		if !empty {
			return fmt.Errorf("目录 %s 非空", s.opts.Dir)
		}
	}
	return nil
}

// resolveVersion 查基底远端的标签，决定 clone 之后要 checkout 到哪个版本。
func resolveVersion(s *session) error {
	if s.opts.Version == "main" {
		fmt.Fprintf(s.out, "    使用基底 main 分支（开发中的最新代码）\n")
		return nil
	}

	probe := runner{dir: ".", out: s.out, verbose: s.opts.Verbose}
	outStr, err := probe.git("ls-remote", "--tags", "--refs", s.opts.BaseURL, "refs/tags/v*")
	if err != nil {
		return fmt.Errorf("读取基底标签: %w", err)
	}

	var tags []string
	for _, line := range strings.Split(outStr, "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		tags = append(tags, strings.TrimPrefix(parts[1], "refs/tags/"))
	}

	if s.opts.Version != "" {
		for _, t := range tags {
			if t == s.opts.Version {
				s.resolvedTag = t
				fmt.Fprintf(s.out, "    使用基底 %s\n", t)
				return nil
			}
		}
		return fmt.Errorf("基底没有标签 %s（可用: %s）", s.opts.Version, strings.Join(tags, ", "))
	}

	latest, ok := latestTag(tags)
	if !ok {
		return fmt.Errorf("基底还没有版本标签，请加 --version main")
	}
	s.resolvedTag = latest
	fmt.Fprintf(s.out, "    使用基底最新版本 %s\n", latest)
	return nil
}

// cloneBase 用 -o base 直接把远端命名为 base（省掉 remote rename），
// --no-tags 避免基底标签污染下游标签空间，并把 tagOpt 持久化到 remote 配置。
func cloneBase(s *session) error {
	parent := filepath.Dir(s.opts.Dir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	r := runner{dir: parent, out: s.out, verbose: s.opts.Verbose}
	if _, err := r.git("clone", "--no-tags", "-o", "base", s.opts.BaseURL, s.opts.Dir); err != nil {
		return err
	}
	s.git = runner{dir: s.opts.Dir, out: s.out, verbose: s.opts.Verbose}
	return nil
}

// fetchBaseTags 把基底标签拉到 base/v* 命名空间，与下游自己的 v* 标签隔离。
func fetchBaseTags(s *session) error {
	_, err := s.git.git("fetch", "--no-tags", "base", "+refs/tags/v*:refs/tags/base/v*")
	return err
}

func checkoutVersion(s *session) error {
	if s.resolvedTag == "" {
		sha, err := s.git.git("rev-parse", "--short", "HEAD")
		if err != nil {
			return err
		}
		s.versionLabel = "main@" + sha
		return nil
	}
	if _, err := s.git.git("reset", "--hard", "refs/tags/base/"+s.resolvedTag); err != nil {
		return err
	}
	s.versionLabel = s.resolvedTag
	return nil
}

// configureRemotes 断开与基底 main 的 upstream 跟踪（否则 git push 会推向基底），
// 并按需添加 origin。不自动 push：远端仓库可能还没建。
func configureRemotes(s *session) error {
	// 没有 upstream 时会报错，忽略即可
	_, _ = s.git.git("branch", "--unset-upstream")
	if s.opts.Origin != "" {
		if _, err := s.git.git("remote", "add", "origin", s.opts.Origin); err != nil {
			return err
		}
		fmt.Fprintf(s.out, "    origin -> %s（尚未 push）\n", s.opts.Origin)
	}
	return nil
}

// writeConfigYAML 生成 server/config.yaml 并换掉占位密钥。
// 生产模式会拒绝占位值和过短的 secret，这里直接给一个够强的。
func writeConfigYAML(s *session) error {
	src := filepath.Join(s.opts.Dir, "server", "config.yaml.example")
	dst := filepath.Join(s.opts.Dir, "server", "config.yaml")
	if err := copyFile(src, dst); err != nil {
		return err
	}
	secret, err := randomHex(32)
	if err != nil {
		return err
	}
	return replaceOnce(dst, `secret: "change-this-to-a-strong-secret"`, `secret: "`+secret+`"`)
}

// writeDeployEnv 生成 .deploy.env。PROJECT_NAME 决定 systemd 服务名、远程目录归属
// 校验和隧道前缀，同机多项目全靠它区分。
func writeDeployEnv(s *session) error {
	src := filepath.Join(s.opts.Dir, ".deploy.env.example")
	dst := filepath.Join(s.opts.Dir, ".deploy.env")
	if err := copyFile(src, dst); err != nil {
		return err
	}
	if err := replaceOnce(dst, "PROJECT_NAME=myproject", "PROJECT_NAME="+s.opts.Name); err != nil {
		return err
	}
	if err := replaceOnce(dst, "REMOTE_SERVER_DIR=/opt/myproject/server", "REMOTE_SERVER_DIR=/opt/"+s.opts.Name+"/server"); err != nil {
		return err
	}
	return replaceOnce(dst, "REMOTE_ADMIN_DIR=/opt/myproject/admin", "REMOTE_ADMIN_DIR=/opt/"+s.opts.Name+"/admin")
}

// writeAdminEnv 改前端 .env（这是被 git 跟踪的文件，改动会进初始化提交）。
// namespace 必须按项目区分：同域名下两个后台共用 namespace 会互相覆盖登录态。
func writeAdminEnv(s *session) error {
	path := filepath.Join(s.opts.Dir, "admin", "apps", "web-antd", ".env")
	if err := replaceOnce(path, "VITE_APP_TITLE=Admin", "VITE_APP_TITLE="+s.opts.Title); err != nil {
		return err
	}
	if err := replaceOnce(path, "VITE_APP_NAMESPACE=go-base-admin", "VITE_APP_NAMESPACE="+s.opts.Name+"-admin"); err != nil {
		return err
	}
	key, err := randomToken(24)
	if err != nil {
		return err
	}
	return replaceOnce(path, "VITE_APP_STORE_SECURE_KEY=please-replace-me-with-your-own-key", "VITE_APP_STORE_SECURE_KEY="+key)
}

// writeChuanyunToml 默认不生成：dev.sh 的 chuanyun_slug 会回落到 .deploy.env 的
// PROJECT_NAME，隧道名照样按项目区分；只有要声明 [[tunnels]]/[[connects]] 才需要这个文件。
func writeChuanyunToml(s *session) error {
	if !s.opts.Chuanyun {
		fmt.Fprintf(s.out, "    跳过（隧道名会自动取 PROJECT_NAME；需要时用 --chuanyun）\n")
		return nil
	}
	src := filepath.Join(s.opts.Dir, "chuanyun.toml.example")
	dst := filepath.Join(s.opts.Dir, "chuanyun.toml")
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return replaceOnce(dst, `project = "base"`, `project = "`+s.opts.Name+`"`)
}

// installDeps 用 go mod download 而不是 tidy：tidy 可能改写 go.sum，
// 让刚创建的项目一上来工作区就是脏的。
func installDeps(s *session) error {
	if s.opts.SkipInstall {
		fmt.Fprintf(s.out, "    跳过\n")
		return nil
	}
	serverRunner := runner{dir: filepath.Join(s.opts.Dir, "server"), out: s.out, verbose: s.opts.Verbose}
	if err := serverRunner.stream("go", "mod", "download"); err != nil {
		return err
	}
	adminRunner := runner{dir: filepath.Join(s.opts.Dir, "admin"), out: s.out, verbose: s.opts.Verbose}
	return adminRunner.stream("pnpm", "install", "--frozen-lockfile")
}

// commitInit 提交初始化改动。--no-verify: admin/lefthook.yml 配了 commitlint，
// 若下游装了 hook，中文提交信息会被 conventional-commit 规则拦下。
func commitInit(s *session) error {
	if _, err := s.git.git("add", "-A"); err != nil {
		return err
	}
	status, err := s.git.git("status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) == "" {
		fmt.Fprintf(s.out, "    无改动可提交（基底模板已是目标状态）\n")
		return nil
	}
	msg := fmt.Sprintf("Initialize %s from base %s", s.opts.Name, s.versionLabel)
	_, err = s.git.git("commit", "--no-verify", "-q", "-m", msg)
	return err
}
