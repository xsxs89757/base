SHELL := /bin/bash

.DEFAULT_GOAL := help

DEPLOY_MODE ?= all
DEV_FLAGS :=

ifeq ($(FORCE),1)
DEV_FLAGS += --force
endif

.PHONY: help dev dev-force force-dev release publish release-server publish-server release-admin publish-admin build build-server build-admin test test-server swagger

help:
	@echo "Admin 管理系统快捷命令"
	@echo ""
	@echo "开发:"
	@echo "  make dev              启动后端 air + 前端 Vite"
	@echo "  make dev-force        强制释放开发端口后启动"
	@echo "  make dev FORCE=1      同 make dev-force"
	@echo ""
	@echo "发布:"
	@echo "  make release          全量发布，等价 ./deploy.sh all"
	@echo "  make release-server   仅发布后端"
	@echo "  make release-admin    仅发布后台前端"
	@echo "  make publish          release 的别名"
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
	@./deploy.sh $(DEPLOY_MODE)

release-server publish-server:
	@./deploy.sh server

release-admin publish-admin:
	@./deploy.sh admin

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
