package main

const PowerShell = "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"
const Store = "store.json"

func main() {
	app := &Application{}
	app.init()
	app.dispatch()
}
