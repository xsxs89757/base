package main

import (
	"fmt"
	"io"
	"os"
)

type step struct {
	name string
	fn   func(*session) error
}

// session 串起所有步骤的共享状态。
type session struct {
	opts options
	out  io.Writer

	// 解析出的基底版本：resolvedTag 为空表示用 main
	resolvedTag string
	// 用于提交信息，main 时是短 SHA
	versionLabel string
	// 目录是本次创建的，失败时才允许删
	createdDir bool
	// 在目标目录里执行命令
	git runner
}

func run(opts options, out io.Writer) error {
	s := &session{opts: opts, out: out}

	steps := []step{
		{"检查环境", preflight},
		{"解析基底版本", resolveVersion},
		{"克隆基底", cloneBase},
		{"拉取基底标签", fetchBaseTags},
		{"切到目标版本", checkoutVersion},
		{"配置远端", configureRemotes},
		{"生成后端配置", writeConfigYAML},
		{"生成部署配置", writeDeployEnv},
		{"设置前端应用标识", writeAdminEnv},
		{"生成穿云配置", writeChuanyunToml},
		{"启用 pre-push 钩子", enableHooks},
		{"安装依赖", installDeps},
		{"提交初始化", commitInit},
	}

	for i, st := range steps {
		fmt.Fprintf(out, "==> [%d/%d] %s\n", i+1, len(steps), st.name)
		if err := st.fn(s); err != nil {
			s.cleanup()
			return fmt.Errorf("%s: %w", st.name, err)
		}
	}

	printNextSteps(s)
	return nil
}

// cleanup 只删本次创建的目录：用户指定了一个已存在的空目录时不该被连锅端。
func (s *session) cleanup() {
	if !s.createdDir || s.opts.KeepOnError {
		if s.createdDir && s.opts.KeepOnError {
			fmt.Fprintf(s.out, "\n已保留目录供排查: %s\n", s.opts.Dir)
		}
		return
	}
	if err := os.RemoveAll(s.opts.Dir); err != nil {
		fmt.Fprintf(s.out, "\n清理目录失败（请手动删除 %s）: %v\n", s.opts.Dir, err)
	}
}

func printNextSteps(s *session) {
	fmt.Fprintf(s.out, "\n完成！项目已创建在 %s（基底 %s）\n\n", s.opts.Dir, s.versionLabel)
	fmt.Fprintf(s.out, "下一步:\n")
	fmt.Fprintf(s.out, "  cd %s\n", s.opts.Dir)
	if s.opts.Origin == "" {
		fmt.Fprintf(s.out, "  git remote add origin <你的仓库地址>\n")
	}
	fmt.Fprintf(s.out, "  git push -u origin main\n")
	fmt.Fprintf(s.out, "  make hooks                   # 启用 pre-push 检查，push 前自动按改动路径验证\n")
	fmt.Fprintf(s.out, "  make dev                     # 启动前后端，默认账号 super / 123456\n")
	fmt.Fprintf(s.out, "\n还需要手工处理:\n")
	fmt.Fprintf(s.out, "  .deploy.env                  填 SSH_HOST / SSH_USER / SSH_PASS 后才能 make release\n")
	fmt.Fprintf(s.out, "  README.md / AGENTS.md        改成本项目自己的说明（CLAUDE.md 只是 @AGENTS.md，不用动）\n")
	fmt.Fprintf(s.out, "  make sync-base               以后合入基底更新；make base-version 查看版本\n")
	if s.opts.SkipInstall {
		fmt.Fprintf(s.out, "\n本次跳过了依赖安装，首次启动前先跑:\n")
		fmt.Fprintf(s.out, "  (cd server && go mod download) && (cd admin && pnpm install)\n")
	}
}
