package goconfig

import (
	"flag"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	ConfigFile *goconfig.ConfigFile

	ROOT string

	configRoot string

	TemplateDir string
)

const mainIniPath = "env.ini"

func init() {

	configFlag := flag.String("config", "", "配置文件路径")
	flag.Parse()

	var err error
	var configPath string
	if (*configFlag != "" ) {
		configPath = *configFlag
		configRoot, err = exec.LookPath(configPath)
		if err != nil {
			panic(err)
		}

		ConfigFile, err = goconfig.LoadConfigFile(configPath)
		if err != nil {
			panic(err)
		}

		if err = loadIncludeFiles(); err != nil {
			panic("load include files error:" + err.Error())
		}

		appConfig := make(map[string]string)
		appConfig, err =ConfigFile.GetSection("app")

		ROOT = appConfig["appPath"]

		configRoot = ROOT + "/config/"

	} else {
		curFilename := os.Args[0]

		var binaryPath string
		binaryPath, err = exec.LookPath(curFilename)
		if err != nil {
			panic(err)
		}

		binaryPath, err = filepath.Abs(binaryPath)
		if err != nil {
			panic(err)
		}

		ROOT = filepath.Dir(filepath.Dir(binaryPath))

		configRoot = ROOT + "/config/"

		configPath = configRoot + mainIniPath

		if !fileExist(configPath) {
			panic("can't find " + mainIniPath)
		}

		ConfigFile, err = goconfig.LoadConfigFile(configPath)
		if err != nil {
			panic(err)
		}

		if err = loadIncludeFiles(); err != nil {
			panic("load include files error:" + err.Error())
		}
	}


	TemplateDir = ROOT + "/template/"

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGUSR1)

		for {
			sig := <-ch
			switch sig {
			case syscall.SIGUSR1:
				ReloadConfigFile()
			}
		}
	}()

}


func ReloadConfigFile() {
	var err error
	configPath := configRoot + mainIniPath
	ConfigFile, err = goconfig.LoadConfigFile(configPath)
	if err != nil {
		fmt.Println("reload config file, error:", err)
		return
	}

	if err = loadIncludeFiles(); err != nil {
		fmt.Println("reload files include files error:", err)
		return
	}
	fmt.Println("reload config file successfully！")
}

func loadIncludeFiles() error {
	includeFile := ConfigFile.MustValue("include_files", "path", "")
	if includeFile != "" {
		includeFiles := strings.Split(includeFile, ",")

		incFiles := make([]string, len(includeFiles))
		for i, incFile := range includeFiles {
			incFiles[i] = configRoot + incFile
		}
		return ConfigFile.AppendFiles(incFiles...)
	}

	return nil
}

// fileExist 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
