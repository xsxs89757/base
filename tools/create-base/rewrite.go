package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 项目名同时要当 systemd 服务名、穿云隧道子域名（DNS label）和部署目录名，
// 所以限制成小写字母开头的 [a-z0-9-]。
var nameRe = regexp.MustCompile(`^[a-z][a-z0-9-]{1,39}$`)

func validateName(name string) error {
	if !nameRe.MatchString(name) {
		return fmt.Errorf("项目名 %q 不合法：需小写字母开头，只含小写字母/数字/连字符，2-40 位", name)
	}
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("项目名 %q 不能以连字符结尾", name)
	}
	if name == "base" {
		return fmt.Errorf("项目名不能叫 base（与基底同名，部署和隧道都会混淆）")
	}
	return nil
}

// replaceOnce 要求 old 在文件中恰好出现一次。
// 模板漂移（占位符被改名或出现多处）时立刻失败，好过静默生成一份错配置。
func replaceOnce(path, old, new string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(raw)
	if n := strings.Count(content, old); n != 1 {
		return fmt.Errorf("%s 中 %q 出现 %d 次（期望 1 次），基底模板可能已变更", filepath.Base(path), old, n)
	}
	return os.WriteFile(path, []byte(strings.Replace(content, old, new, 1)), 0o644)
}

// copyFile 用于从 *.example 生成实际配置文件。
func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, raw, 0o644)
}

// randomHex 生成 n 字节的十六进制串（jwt.secret 用，长度 2n）。
func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// randomToken 生成 URL 安全的随机串（前端 store 加密密钥用）。
func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
