package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
)

var d = flag.String("d", ".", "监听的目录dir")

func main() {

	flag.Parse()

	var args = flag.Args()

	if len(args) > 0 {
		conf.Command = args
	} else {
		log.Println("请输入命令")
	}

	conf.WatchDirs = append([]string{}, *d)

	log.Println("监听目录:", *d)

	log.Println("执行命令:", args)

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
var cancel context.CancelFunc

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

	exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(cmd.Process.Pid)).Run()
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
