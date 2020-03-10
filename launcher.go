package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Title       string   `yaml:"title"`
	Port        int      `yaml:"port"`
	ItemList    []Item   `yaml:"items"`
	Directories []string `yaml:"directories"`
}

// structure to hold each launcher item
type Item struct {
	Title       string   `yaml:"title"`
	Enable      bool     `yaml:"enable"`
	Description string   `yaml:"description"`
	Authors     []string `yaml:"authors"`
	Command     string   `yaml:"command"`
	Hosts       []string `yaml:"hosts"`
	LocalDir    string
	Handle      *exec.Cmd
	ID          int
}

var globalConfig *Config

func save(li *Item) error {
	filename := li.Title + ".yaml"
	return ioutil.WriteFile(filename, []byte(li.Description), 0600)
}

func loadGlobalConfig(file string) (err error) {

	globalConfig = &Config{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, globalConfig)

	if err != nil {
		fmt.Println(err)
	}

	return
}

func updateLauncherItems() {

	// compute hashes for launch id:
	for i := range globalConfig.ItemList {
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

	if err != nil {
		fmt.Println(err)
	}

	item.LocalDir = filepath.Dir(file)

	globalConfig.ItemList = append(globalConfig.ItemList, *item)

	return nil
}

func loadBootStrap() {

	// load global config
	err := loadGlobalConfig("data/items.yml")
	if err != nil {
		fmt.Println("failed to load config ", err)
	}

	// now search subdirectories
	for _, dir := range globalConfig.Directories {

		fmt.Println(dir)

		err := filepath.Walk(dir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				matched, _ := filepath.Match("gillaunch.yml", filepath.Base(path))
				fmt.Println(path, matched)
				if matched {
					loadLauncher(path)
					fmt.Println(path, info.Size())
				}
				return nil
			})

		if err != nil {
			log.Println(err)
		}
	}
	// update IDs
	updateLauncherItems()

}

func reloadHandler(w http.ResponseWriter, r *http.Request) {

	loadBootStrap()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func killHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		q := u.Query()

		idx, err := strconv.ParseInt(q["id"][0], 10, 64)

		if err != nil {
			if globalConfig.ItemList[idx].Handle != nil && globalConfig.ItemList[idx].Handle.Process != nil {
				err = globalConfig.ItemList[idx].Handle.Process.Kill()
			}
		}

		if err != nil {
			log.Fatal(err)
		}

		globalConfig.ItemList[idx].Handle = nil
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (item *Item) launch() {

	fullCommand := path.Join(item.LocalDir, item.Command)

	fmt.Println("launching ", fullCommand)

	item.Handle = exec.Command(fullCommand)

	item.Handle.Start()
}

func launchHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Fatal(err)
		}

		q := u.Query()

		idx, err := strconv.ParseInt(q["id"][0], 10, 64)

		globalConfig.ItemList[idx].launch()

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

	// load everything
	loadBootStrap()

	// hoist server
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// http handler
	http.HandleFunc("/", handler)
	http.HandleFunc("/launch", launchHandler)
	http.HandleFunc("/kill", killHandler)
	http.HandleFunc("/reload", reloadHandler)

	// start stuff
	port := strconv.Itoa(globalConfig.Port)

	log.Fatal(http.ListenAndServe(":"+port, nil))

}
