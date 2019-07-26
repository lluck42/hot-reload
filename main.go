package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

var dir = flag.String("dir", "", "设置fs监听目录")

func main() {
	// 载入配置
	getConfig()
	// 命令行配置优先
	flag.Parse()

	var args = flag.Args()
	if len(args) > 0 {
		conf.Command = args
	}
	if *dir != "" {
		conf.WatchDirs = append([]string{}, *dir)
	}
	log.Println(conf)

	start()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 递归设置的目录
	// 加入监听
	for _, v := range conf.WatchDirs {
		d, err := GetAllDir(v)
		if err != nil {
			panic(err)
		}
		for _, vv := range d {
			if watcher.Add(vv) != nil {
				log.Fatal(err)
			}
		}
	}

	// 处理监听事件
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fmt.Println(event.Op)
			restart()
		}
	}

	// 不退出代码
	// done := make(chan bool)
	// <-done
}

var cmd *exec.Cmd

// 启动
func start() {

	cmd = exec.Command(conf.Command[0], conf.Command[1:]...)
	// cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}

// kill
func kill() {

	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			fmt.Println(err)
		}
		cmd.Wait()
	}
}

// restart
func restart() {
	kill()
	start()
}

// Config Config
type Config struct {
	Command   []string
	WatchDirs []string
}

var conf Config

// getConfig
func getConfig() {
	var exis, _ = PathExists("./config.yaml")
	if !exis {
		return
	}
	yamlFile, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Println(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Println(err.Error())
	}
}

// GetAllDir GetAllDir
func GetAllDir(pathname string) ([]string, error) {
	var allDir []string
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return allDir, err
	}

	allDir = append(allDir, pathname)

	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			subdir, err := GetAllDir(fullDir)
			allDir = append(allDir, subdir...)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return allDir, err
			}
		}
	}
	return allDir, nil
}

// PathExists PathExists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
