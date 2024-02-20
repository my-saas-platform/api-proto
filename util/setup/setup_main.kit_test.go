package setuputil

import (
	"os"
	"testing"
)

// confPath 配置目录
const confPath = "./../../testdata/configs"

func TestMain(m *testing.M) {

	// 初始化必要逻辑
	os.Exit(m.Run())
}
