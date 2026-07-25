package watch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

// TestMain 在运行本包测试前，将工作目录切到模块根目录并加载邮件模板，
// 使依赖 emailtmpl.GenEmailMsg 的测试用例能够渲染模板（Init 读取 ./conf/email-template/*.tmpl）。
// 模板加载失败仅记录、不中断，不影响模板无关的用例。
func TestMain(m *testing.M) {
	if root := findModuleRoot(); root != "" {
		_ = os.Chdir(root)
	}

	if err := emailtmpl.Init(); err != nil {
		// 不 fail 整个包，仅记录；模板相关用例会自行暴露问题。
		os.Exit(m.Run())
	}

	os.Exit(m.Run())
}

// findModuleRoot 从本测试源码所在目录向上查找包含 go.mod 的目录。
func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
