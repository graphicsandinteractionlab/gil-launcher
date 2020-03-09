package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Title    string `yaml:"title"`
	Port     int    `yaml:"port"`
	ItemList []Item `yaml:"items"`
	Directories []string `yaml:"directories"`
}

type Item struct {
	Title       string   `yaml:"title"`
	Enable      bool     `yaml:"enable"`
	Description string   `yaml:"description"`
	Authors     []string `yaml:"authors"`
	Command     string   `yaml:"command"`
	Handle      *exec.Cmd
	ID          int
}

var globalConfig = &Config{}

func save(li *Item) error {
	filename := li.Title + ".yaml"
	return ioutil.WriteFile(filename, []byte(li.Description), 0600)
}

func loadGlobalConfig(file string) (err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, globalConfig)
	return
}

func updateLauncherItems() {

	// compute hashes for launch id:
	for i := range globalConfig.ItemList {
		fmt.Println(i)
		globalConfig.ItemList[i].ID = i
	}
}

func loadLauncher(file string) (err error) {

	item := &Item{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, item)

	globalConfig.ItemList = append(globalConfig.ItemList, *item)

	return nil
}

func killHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		q := u.Query()

		idx, err := strconv.ParseInt(q["id"][0], 10, 64)

		err = globalConfig.ItemList[idx].Handle.Process.Kill()

		if err != nil {
			log.Fatal(err)
		}

		globalConfig.ItemList[idx].Handle = nil
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func launchHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		q := u.Query()

		idx, err := strconv.ParseInt(q["id"][0], 10, 64)

		commandline := globalConfig.ItemList[idx].Command

		globalConfig.ItemList[idx].Handle = exec.Command(commandline)

		globalConfig.ItemList[idx].Handle.Start()

	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		fmt.Println(r.Form)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	tmpl, _ := template.ParseFiles("templates/view.html")

	tmpl.Execute(w, globalConfig)
}

func main() {

	// load global config
	err := loadGlobalConfig("data/items.yml")
	if err != nil {
		fmt.Println("failed to load config ", err)
	}

	// now search subdirectories
	for _,dir := range globalConfig.Directories {
		launcherFiles, _ := filepath.Glob(dir + "/*/gillaunch.yml")
		for _, item := range launcherFiles {
			loadLauncher(item)
		}	
	}

	// update IDs
	updateLauncherItems()

	for _, item := range globalConfig.ItemList {
		fmt.Println("--")
		fmt.Println(item)
	}

	// hoist server
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handler)
	http.HandleFunc("/launch", launchHandler)
	http.HandleFunc("/kill", killHandler)

	log.Fatal(http.ListenAndServe(":8181", nil))

}
