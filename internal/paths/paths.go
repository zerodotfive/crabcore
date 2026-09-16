package paths

import (
	"os"
	"path/filepath"
)

func Home() string {
	return os.Getenv("HOME")
}

func LocalConfigFile() string {
	return filepath.Join(Home(), ".crabcore.yaml")
}

func BinDir() string {
	return filepath.Join(Home(), ".local/bin")
}

func BinInstallPath() string {
	return filepath.Join(BinDir(), "crabcore")
}

func SystemdUserDir() string {
	return filepath.Join(Home(), ".config/systemd/user")
}

func LaunchAgentDir() string {
	return filepath.Join(Home(), "Library/LaunchAgents")
}

func Bashrc() string {
	return filepath.Join(Home(), ".bashrc")
}

func Zshrc() string {
	return filepath.Join(Home(), ".zshrc")
}

func PluginLibDir() string {
	return filepath.Join(Home(), ".local/share/crabcore/lib")
}

func PluginLibPath(filename string) string {
	return filepath.Join(PluginLibDir(), filename)
}

func PluginConfigDir(pluginName string) string {
	return filepath.Join(Home(), ".local/share/crabcore/config", pluginName)
}

func ModuleConfigPath(pluginName, moduleName string) string {
	return filepath.Join(PluginConfigDir(pluginName), moduleName+".yaml")
}
