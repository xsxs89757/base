SHELL := /bin/bash

.DEFAULT_GOAL := help

DEPLOY_MODE ?= all
PROJECT ?=
DEV_FLAGS :=

# 基底仓库标识：base-release 用它确认「只在基底本体发布」
BASE_REPO := xsxs89757/base
BASE_URL := https://github.com/xsxs89757/base.git
# 版本号：base-release 必填 (vX.Y.Z)；sync-base 可选 (vX.Y.Z | main)
VERSION ?=
# swag 钉死版本：本地与 CI 必须用同一版本，否则 docs 一致性检查会误报
SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6

ifeq ($(FORCE),1)
DEV_FLAGS += --force
endif

# 下游挂载点：基底永不提供 Makefile.project，下游在其中写自己的目标，同步永不冲突。
# 目标带 `## 说明` 注释就会出现在 make help 里；PROJECT_CHECKS 里列出的目标会并入 make check。
-include Makefile.project

.PHONY: help dev dev-force force-dev release publish release-server publish-server release-admin publish-admin build build-server build-admin test test-server swagger sync-base check-hooks base-check base-release base-version new migrate-kit kit-dev kit-undev
.PHONY: check check-backend check-frontend check-scripts typecheck hooks

help:
	@echo "Admin 管理系统快捷命令"
	@echo ""
	@echo "开发:"
	@echo "  make dev              启动后端 air + 前端 Vite (端口被占用自动改用空闲端口)"
	@echo "  make dev-force        杀死占用进程，坚持使用配置端口启动"
	@echo "  make dev FORCE=1      同 make dev-force"
	@echo ""
	@echo "发布(部署到服务器):"
	@echo "  make release          全量发布，等价 ./deploy.sh all (含下游扩展目标)"
	@echo "  make release-server   仅发布后端"
	@echo "  make release-admin    仅发布后台前端"
	@echo "  make release-<目标>   发布 deploy.project.sh 声明的下游扩展目标"
	@echo "  make publish          release 的别名"
	@echo ""
	@echo "  多项目: make release PROJECT=shop   使用 .deploy.shop.env 发布"
	@echo "          ./deploy.sh --list          查看已有部署配置"
	@echo ""
	@echo "基底:"
	@echo "  make sync-base        下游合入基底最新版本 (默认最新标签)"
	@echo "  make sync-base VERSION=v1.2.0   合入指定版本; VERSION=main 合入开发中的 main"
	@echo "  make base-version     查看当前已合入的基底版本与远端最新版本"
	@echo ""
	@echo "基底维护(仅基底仓库本体):"
	@echo "  make check-hooks      校验下游挂载点未被基底改动"
	@echo "  make base-check       发布前全量检查 (挂载点/测试/构建/脚本语法/Swagger 时效)"
	@echo "  make base-release VERSION=vX.Y.Z   打版本标签并推送"
	@echo "  make new NAME=demo    用本地代码跑一遍脚手架 (开发脚手架时用)"
	@echo ""
	@echo "框架层 (base-kit):"
	@echo "  make migrate-kit      从 v2.0.0 之前的版本同步后，改写 import 路径"
	@echo "  make kit-dev          用本地 ../base-kit 源码开发框架层 (生成 server/go.work)"
	@echo "  make kit-undev        改回按 go.mod 钉死的版本"
	@echo ""
	@echo "验证/构建:"
	@echo "  make check            跑一遍 CI 的全部检查 (后端 + 前端 + 脚本)"
	@echo "  make check-backend    仅后端 (vet/test/交叉编译/Swagger)"
	@echo "  make check-frontend   仅前端 (类型检查 + 构建)"
	@echo "  make typecheck        仅前端类型检查 (最快)"
	@echo "  make hooks            启用 .githooks (push 前自动按改动路径检查)"
	@echo "  make test             运行后端测试"
	@echo "  make build            构建后端和前端"
	@echo "  make swagger          重新生成 Swagger 文档"
	@[ ! -f Makefile.project ] || { echo ""; echo "项目自有目标 (Makefile.project):"; \
		grep -hE '^[a-zA-Z0-9_-]+:.*## ' Makefile.project | \
		awk -F':.*## ' '{printf "  make %-18s %s\n", $$1, $$2}'; }

dev:
	@./dev.sh $(DEV_FLAGS)

dev-force force-dev:
	@./dev.sh --force

release publish:
	@./deploy.sh $(DEPLOY_MODE) $(PROJECT)

release-server publish-server:
	@./deploy.sh server $(PROJECT)

release-admin publish-admin:
	@./deploy.sh admin $(PROJECT)

# 下游扩展目标 (deploy.project.sh 的 PROJECT_DEPLOY_TARGETS)，如 make release-agent
release-%:
	@./deploy.sh $* $(PROJECT)

# 框架层在 github.com/xsxs89757/base-kit，用 go get github.com/xsxs89757/base-kit@latest 升级；
# 下面三个目标只在少数场景用到
migrate-kit:
	@echo "==> 改写 import 路径 (base/internal/* -> base-kit)"
	@cd server && go run github.com/xsxs89757/base-kit/cmd/basekit-migrate ./... || { \
		echo "跑不起来通常是 go.mod 里还没有 base-kit：先 make sync-base，"; \
		echo "解决 go.mod 冲突时保留 require github.com/xsxs89757/base-kit 那一行"; exit 1; }
	@cd server && go mod tidy
	@echo "==> 完成，接着跑 make swagger && make test"

# 同时改 kit 和模板时用：go.work 让模板直接编译 ../base-kit 的源码，改完 kit 不用发版就能验证。
# go.work 已 gitignore；deploy.sh 用 GOWORK=off 编译，发布永远按 go.mod 钉死的版本。
kit-dev:
	@[ -d ../base-kit ] || { echo "未找到 ../base-kit，先 git clone https://github.com/xsxs89757/base-kit.git"; exit 1; }
	@cd server && go work init . ../../base-kit 2>/dev/null || true
	@cd server && go list -m -f '  base-kit 现在解析到 {{.Dir}}' github.com/xsxs89757/base-kit

kit-undev:
	@rm -f server/go.work server/go.work.sum
	@cd server && go list -m -f '  base-kit 现在解析到 {{.Dir}}' github.com/xsxs89757/base-kit

# ---------------------------------------------------------------------------
# 校验：这些只是 scripts/check.sh 的转发，CI 与 pre-push 钩子调的是同一份逻辑。
# 改检查内容请改 scripts/check.sh，不要在这里加命令。
# ---------------------------------------------------------------------------

check:
	@bash scripts/check.sh all
	@[ -z "$(PROJECT_CHECKS)" ] || $(MAKE) --no-print-directory $(PROJECT_CHECKS)

check-backend:
	@bash scripts/check.sh backend

check-frontend:
	@bash scripts/check.sh frontend

check-scripts:
	@bash scripts/check.sh scripts

typecheck:
	@bash scripts/check.sh frontend --fast

# 启用 pre-push 钩子。dev.sh 启动时也会自动调用（仅在未设置 core.hooksPath 时）。
hooks:
	@git config core.hooksPath .githooks
	@echo "已启用 .githooks（关闭: git config --unset core.hooksPath）"

build: build-server build-admin

build-server:
	@cd server && go build ./...

build-admin:
	@cd admin && pnpm build:antd

test: test-server

test-server:
	@cd server && go test ./...

swagger:
	@cd server && $(SWAG) init -g main.go -o docs --parseDependencyLevel 3 --packagePrefix base,github.com/xsxs89757/base-kit

# ---------------------------------------------------------------------------
# 基底与下游同步
# ---------------------------------------------------------------------------

# 基底标签在下游以 base/v* 命名空间存在，避免与下游自己的 v* 标签冲突。
# tagOpt=--no-tags 是必须的：裸 git fetch 会自动跟进基底标签，污染下游标签空间。
sync-base:
	@git remote get-url base >/dev/null 2>&1 || { \
		echo "未找到名为 base 的 remote（基底仓库本体无需同步）。"; \
		echo "下游项目请先执行: git remote add base $(BASE_URL)"; \
		exit 1; }
	@git config remote.base.tagOpt --no-tags
	@git fetch --no-tags base '+refs/heads/*:refs/remotes/base/*' '+refs/tags/v*:refs/tags/base/v*'
	@ref="$(VERSION)"; \
	if [ -z "$$ref" ]; then \
		ref=$$(git tag -l 'base/v*' --sort=-v:refname | grep -E '^base/v[0-9]+\.[0-9]+\.[0-9]+$$' | head -1); \
		[ -n "$$ref" ] || { echo "基底还没有版本标签，请用: make sync-base VERSION=main"; exit 1; }; \
	elif [ "$$ref" = main ]; then \
		ref=base/main; \
	elif echo "$$ref" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		ref="base/$$ref"; \
		git rev-parse -q --verify "refs/tags/$$ref" >/dev/null || { echo "基底没有标签 $(VERSION)"; exit 1; }; \
	else \
		echo "VERSION 须为 vX.Y.Z 或 main"; exit 1; \
	fi; \
	echo "==> 当前已合入: $$(cat .base-version 2>/dev/null || echo '未知(早于 v1.0.0)')  ->  合入 $$ref"; \
	git merge "$$ref"

base-version:
	@echo "当前已合入的基底版本: $$(cat .base-version 2>/dev/null || echo '未知(早于 v1.0.0)')"
	@git remote get-url base >/dev/null 2>&1 || { echo "(此处是基底仓库本体)"; exit 0; }; \
	git fetch -q --no-tags base '+refs/tags/v*:refs/tags/base/v*' 2>/dev/null || true; \
	latest=$$(git tag -l 'base/v*' --sort=-v:refname | grep -E '^base/v[0-9]+\.[0-9]+\.[0-9]+$$' | head -1); \
	[ -n "$$latest" ] && echo "基底最新版本:         $${latest#base/}" || echo "基底最新版本:         (无标签)"; \
	for t in $$(git tag -l 'base/v*' --sort=-v:refname); do \
		if git merge-base --is-ancestor "$$t" HEAD 2>/dev/null; then echo "历史中已包含:         $${t#base/}"; break; fi; \
	done; true

# ---------------------------------------------------------------------------
# 基底维护（只在基底仓库本体有意义）
# ---------------------------------------------------------------------------

check-hooks:
	@bash scripts/check-hooks.sh

# 发布前全量检查 = 所有人都跑的 check + 只有基底本体才有意义的那部分
base-check:
	@bash scripts/check.sh all
	@bash scripts/check.sh base
	@echo "==> base-check 通过"

base-release:
	@[ -n "$(VERSION)" ] || { echo "用法: make base-release VERSION=vX.Y.Z"; exit 1; }
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "版本号须形如 vX.Y.Z"; exit 1; }
	@git remote get-url origin | grep -Eq '$(BASE_REPO)(\.git)?/?$$' || { echo "只能在基底仓库本体执行（origin 需指向 $(BASE_REPO)）"; exit 1; }
	@[ "$$(git symbolic-ref --short HEAD)" = main ] || { echo "须在 main 分支"; exit 1; }
	@[ -z "$$(git status --porcelain)" ] || { echo "工作区不干净，先提交或清理"; exit 1; }
	@git fetch -q origin
	@[ "$$(git rev-parse HEAD)" = "$$(git rev-parse origin/main)" ] || { echo "本地 main 与 origin/main 不一致：先 push 内容提交并等 CI 通过"; exit 1; }
	@! git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null || { echo "标签 $(VERSION) 已存在"; exit 1; }
	@grep -Eq '^## \[$(patsubst v%,%,$(VERSION))\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$$' CHANGELOG.md || { \
		echo "CHANGELOG.md 缺少条目: ## [$(patsubst v%,%,$(VERSION))] - YYYY-MM-DD"; exit 1; }
	@$(MAKE) --no-print-directory base-check
	@printf '%s\n' "$(VERSION)" > .base-version
	@git add .base-version && git commit -q -m "Release $(VERSION)"
	@git tag -a "$(VERSION)" -m "base $(VERSION)"
	@tool_tag=""; \
	last=$$(git tag -l 'tools/create-base/v*' --sort=-v:refname | head -1); \
	if [ -z "$$last" ] || ! git diff --quiet "$$last" HEAD -- tools/create-base; then \
		git tag -a "tools/create-base/$(VERSION)" -m "create-base $(VERSION)"; \
		tool_tag="refs/tags/tools/create-base/$(VERSION)"; \
		echo "==> 已打脚手架标签 tools/create-base/$(VERSION)"; \
	else \
		echo "==> tools/create-base 无变化，不打脚手架标签（go run @latest 仍解析到 $$last）"; \
	fi; \
	git push --atomic origin main "refs/tags/$(VERSION)" $$tool_tag
	@echo "==> 已发布 $(VERSION)"

NEW_DIR ?= ../$(NAME)
new:
	@[ -n "$(NAME)" ] || { echo "用法: make new NAME=myproject [VERSION=main] [NEW_DIR=../myproject]"; exit 1; }
	@dir="$(NEW_DIR)"; case "$$dir" in /*) ;; *) dir="$(CURDIR)/$$dir";; esac; \
	cd tools/create-base && go run . --base "$(CURDIR)" --version "$(if $(VERSION),$(VERSION),main)" --dir "$$dir" --skip-install "$(NAME)"
