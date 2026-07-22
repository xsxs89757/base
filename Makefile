SHELL := /bin/bash

.DEFAULT_GOAL := help

DEPLOY_MODE ?= all
PROJECT ?=
DEV_FLAGS :=

ifeq ($(FORCE),1)
DEV_FLAGS += --force
endif

.PHONY: help dev dev-force force-dev release publish release-server publish-server release-admin publish-admin build build-server build-admin test test-server swagger sync-base

help:
	@echo "Admin 管理系统快捷命令"
	@echo ""
	@echo "开发:"
	@echo "  make dev              启动后端 air + 前端 Vite (端口被占用自动改用空闲端口)"
	@echo "  make dev-force        杀死占用进程，坚持使用配置端口启动"
	@echo "  make dev FORCE=1      同 make dev-force"
	@echo ""
	@echo "发布:"
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
	@echo "  make sync-base        下游项目合入基底更新 (git fetch base && git merge base/main)"
	@echo ""
	@echo "验证/构建:"
	@echo "  make test             运行后端测试"
	@echo "  make build            构建后端和前端"
	@echo "  make swagger          重新生成 Swagger 文档"

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

build: build-server build-admin

build-server:
	@cd server && go build ./...

build-admin:
	@cd admin && pnpm build:antd

test: test-server

test-server:
	@cd server && go test ./...

swagger:
	@cd server && swag init -g main.go -o docs --parseDependency --parseInternal

sync-base:
	@git remote get-url base >/dev/null 2>&1 || { \
		echo "未找到名为 base 的 remote（基底仓库本体无需同步）。"; \
		echo "下游项目请先执行: git remote add base https://github.com/xsxs89757/base.git"; \
		exit 1; }
	@git fetch base
	@git merge base/main
