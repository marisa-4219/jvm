package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Application struct {
	PowerShellPath  string            `json:"powershell"`
	SDKPath         string            `json:"sdk_path"`
	SDK             map[string]string `json:"sdk"`
	VirtualJavaHome string            `json:"virtual_java_home"`

	runtime string
	args    []string
}

func (app *Application) init() {
	// get run path
	path, err := os.Executable()
	if err != nil {
		app.exit("unknown error :" + err.Error())
	}
	app.runtime = filepath.Dir(path)

	// read store
	_, err = os.Stat(filepath.Join(app.runtime, Store))
	if os.IsNotExist(err) {
		_, err := os.Create(filepath.Join(app.runtime, Store))
		if err != nil {
			app.exit("cannot create store.json : " + err.Error())
		}
	} else if err != nil {
		app.exit("read store file error :" + err.Error())
	}
	data, err := os.ReadFile(filepath.Join(app.runtime, Store))
	if err != nil {
		app.exit("read store file error :" + err.Error())
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, app)
		if err != nil {
			app.exit("parse store file error :" + err.Error())
		}
	}
	// set default value
	if app.PowerShellPath == "" {
		app.PowerShellPath = PowerShell
	}
	if app.SDK == nil {
		app.SDK = make(map[string]string)
	}
	if app.VirtualJavaHome == "" {
		app.VirtualJavaHome = filepath.Join(app.runtime, "VirtualJavaSDK")
	}
}

func (app *Application) dispatch() {
	var command = ""
	if len(os.Args) == 1 {
		command = "help"
	} else {
		command = strings.ToLower(os.Args[1])
		app.args = os.Args[2:]
	}

	switch command {
	case "install":
		app.install()
	case "use":
		app.use()
	case "set":
		app.set()
	case "store":
		app.store()
	case "search":
		app.search()
	case "update":
		app.update()
	case "help":
		fmt.Println(`
install - install jvm to env
use     - <jdk version> change jdk
set     - <property> <value> set store.json property [sdk_path|virtual_java_home]
store   - show store
search  - list sdk_path all JDK
update  - save sdk_path all JDK to store.json
help    - show help`)
	default:
		app.exit("invalid command :" + os.Args[1])
	}
}

func (app *Application) install() {
	result, err := app.readEnv("Path")
	if err != nil {
		app.exit("read env error :" + err.Error())
	}
	exists := false
	paths := strings.Split(result, ";")

	runtime, _ := os.Stat(app.runtime)

	for i := range paths {
		el, err := os.Stat(paths[i])
		if err != nil {
			continue
		}
		if os.SameFile(runtime, el) {
			exists = true
		}
	}

	if !exists {
		err = app.saveEnv("Path_Backup", result)
		if err != nil {
			app.exit("set env error :" + err.Error())
		}
		err = app.saveEnv("Path", result+";"+app.runtime)
		if err != nil {
			app.exit("set env error :" + err.Error())
		}
	} else {
		app.exit("No need to install anymore")
	}

	app.exit("Install successful, Please restart you are terminal")
}

func (app *Application) use() {
	sdk := app.SDK[app.args[0]]
	if sdk == "" {
		app.exit("sdk not found :" + app.args[0])
	}
	app.symlinks(sdk)
	app.env()

	app.exit("Successful, JDK changed but it may be necessary Please restart you are terminal")
}

func (app *Application) set() {
	if len(app.args) < 2 {
		app.exit("invalid command")
	}
	property := app.args[0]
	value := app.args[1]

	switch strings.ToLower(property) {
	case "sdk_path":
		_, err := os.Stat(value)
		if err != nil {
			app.exit("check path failed :" + err.Error())
		}
		app.SDKPath = value
		app.save()
	case "virtual_java_home":
		_, err := os.Stat(value)
		if err != nil {
			app.exit("check path failed :" + err.Error())
		}
		_, err = os.Stat(filepath.Join(value, "../"))
		if err != nil {
			app.exit("check path failed :" + err.Error())
		}
		app.VirtualJavaHome = value
		app.save()
	default:
		app.exit("invalid property :" + property)
	}

	app.store()
}

func (app *Application) store() {
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		app.exit("parse store failed :" + err.Error())
	}
	fmt.Println("Store:")
	fmt.Println(string(data))
}

func (app *Application) search() {
	sdk := app.traversal()
	if len(sdk) == 0 {
		app.exit("sdk path is empty")
	}
	m := 0
	for s := range sdk {
		if m < len(s) {
			m = len(s)
		}
	}
	for p := range sdk {
		fmt.Println(fmt.Sprintf("%-"+strconv.Itoa(m)+"s", p) + " -> " + sdk[p])
	}
}

func (app *Application) update() {
	sdk := app.traversal()
	if len(sdk) == 0 {
		app.exit("sdk path is empty")
	}
	app.SDK = sdk

	app.save()
	app.exit("Successful")
}

func (app *Application) save() {
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		app.exit("save store error :" + err.Error())
	}
	err = os.WriteFile(filepath.Join(app.runtime, Store), data, 0777)
	if err != nil {
		app.exit("save store error :" + err.Error())
		return
	}
}
func (app *Application) exit(message string) {
	fmt.Println(message)
	os.Exit(0)
}

func (app *Application) readEnv(name string) (string, error) {
	cmd := exec.Command(app.PowerShellPath, "-Command", fmt.Sprintf("[System.Environment]::GetEnvironmentVariable( '%s', 'Machine')", name))
	buf := bytes.Buffer{}
	cmd.Stderr = &buf
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\r\n"), nil
}

func (app *Application) saveEnv(name string, value string) error {
	cmd := exec.Command(app.PowerShellPath, "-Command", fmt.Sprintf("[System.Environment]::SetEnvironmentVariable('%s', '%s', 'Machine')", name, value))
	return cmd.Run()
}

func (app *Application) env() {

	group := sync.WaitGroup{}
	group.Add(3)

	go func() {
		defer group.Done()
		// check path
		result, err := app.readEnv("Path")
		if err != nil {
			app.exit("read env error :" + err.Error())
		}
		exists := false
		paths := strings.Split(result, ";")
		for i := range paths {
			if strings.ToLower(paths[i]) == `%java_home%\bin` {
				exists = true
				break
			}
		}
		if !exists {
			err = app.saveEnv("Path_Backup", result)
			if err != nil {
				app.exit("set env error :" + err.Error())
			}
			err = app.saveEnv("Path", result+`;%JAVA_HOME%\bin`)
			if err != nil {
				app.exit("set env error :" + err.Error())
			}
		}
	}()
	go func() {
		defer group.Done()
		// check classpath
		result, err := app.readEnv("CLASSPATH")
		if err != nil {
			app.exit("read env error :" + err.Error())
		}
		if result == "" || !strings.Contains(result, `%JAVA_HOME%`) {
			err = app.saveEnv("CLASSPATH", `.;%JAVA_HOME%\lib\dt.jar;%JAVA_HOME%\lib\tools.jar;`)
			if err != nil {
				app.exit("set env error :" + err.Error())
			}
		}
	}()
	go func() {
		defer group.Done()
		// check java_home
		err := app.saveEnv("JAVA_HOME", app.VirtualJavaHome)
		if err != nil {
			app.exit("set env error :" + err.Error())
		}
	}()
	group.Wait()
}

func (app *Application) symlinks(targetPath string) {

	err := os.Remove(app.VirtualJavaHome)
	if err != nil && !os.IsNotExist(err) {
		app.exit("create symlinks failed :" + err.Error())
	}

	var symlinks = `New-Item -Path "${linkPath}" -ItemType SymbolicLink -Value "${targetPath}"`
	symlinks = strings.ReplaceAll(symlinks, "${linkPath}", app.VirtualJavaHome)
	symlinks = strings.ReplaceAll(symlinks, "${targetPath}", targetPath)

	err = exec.Command(app.PowerShellPath, "-Command", symlinks).Run()
	if err != nil {
		app.exit("create symlinks failed :" + err.Error())
	}
}

func (app *Application) traversal() map[string]string {
	if app.SDKPath == "" {
		app.exit("Please set sdk_path")
	}
	var task func(group *sync.WaitGroup, name string, path string, callback func(version string, filepath string))

	var load = func(command string) (string, error) {
		cmd := exec.Command(app.PowerShellPath, "-Command", command)
		buf := bytes.Buffer{}
		cmd.Stderr = &buf
		cmd.Stdout = &buf
		err := cmd.Run()
		if err != nil {
			return "", err
		}
		return strings.TrimSuffix(buf.String(), "\r\n"), nil
	}

	task = func(group *sync.WaitGroup, name string, path string, callback func(version string, filepath string)) {
		defer group.Done()
		entries, err := os.ReadDir(path)
		if err != nil {
			app.exit("read dir error :" + err.Error())
		}
		for i := range entries {
			el := entries[i]
			if el.IsDir() {
				group.Add(1)
				go task(group, name, filepath.Join(path, el.Name()), callback)
				continue
			}
			if el.Name() == name {
				version, err := load(filepath.Join(path, el.Name()) + " -version")
				if err != nil {
					app.exit("read dir error :" + err.Error())
				}
				callback(strings.ReplaceAll(version, "javac ", ""), filepath.Join(path, "../"))
			}
		}
	}

	sdk := make(map[string]string)

	group := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	group.Add(1)

	go task(group, "javac.exe", app.SDKPath, func(version string, filepath string) {
		mutex.Lock()
		defer mutex.Unlock()
		sdk[version] = filepath
	})

	group.Wait()

	return sdk
}
