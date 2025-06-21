package file

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetRuntimeDir 返回运行时目录
func GetRuntimeDir() string {
	currentRuntime := "runtime"
	if runtime.GOOS == "darwin" {
		// macOS: 使用 ~/Library/Application Support/noctua/
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join("./", currentRuntime)
		}
		runtimeDir := filepath.Join(home, "Library", "Application Support", "noctua", currentRuntime)
		if err := os.MkdirAll(runtimeDir, 0755); err != nil {
			return filepath.Join("./", currentRuntime)
		}
		return runtimeDir
	}
	return filepath.Join("./", currentRuntime)
}

// 判断是否在 Mac 的 `.app` 包内运行
func isMacAppBundle() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	// 解析符号链接以获取真实路径
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath // 使用未解析的路径
	}
	return strings.Contains(realPath, ".app/Contents/MacOS/")
}

// 获取资源文件的路径
func GetResourcePath(subpath string) string {
	if isMacAppBundle() {
		exePath, err := os.Executable()
		if err != nil {
			return subpath // 回退到相对路径
		}
		realPath, err := filepath.EvalSymlinks(exePath)
		if err != nil {
			realPath = exePath
		}
		// 从 MacOS/ 回退到 Contents/，然后拼接 Resources/
		contentsPath := filepath.Dir(filepath.Dir(realPath)) // 获取 .app/Contents/
		resourcePath := filepath.Join(contentsPath, "Resources", subpath)

		return resourcePath
	}
	if runtime.GOOS == "windows" {
		// 开发环境或 Windows：相对于可执行文件所在目录
		exePath, err := os.Executable()
		if err != nil {
			return subpath
		}
		exeDir := filepath.Dir(exePath)
		resourcePath := filepath.Join(exeDir, subpath)

		return resourcePath
	}

	return filepath.Join("./", subpath)
}
