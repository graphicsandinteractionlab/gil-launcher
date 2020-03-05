package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Title    string `yaml:"title"`
	ItemList []Item `yaml:"items"`
}

type Item struct {
	Enable      bool     `yaml:"enable"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Authors     []string `yaml:"authors"`
	Command     string   `yaml:"command"`
	Handle      *exec.Cmd
	Id          int
}

var globalConfig = &Config{}

func save(li *Item) error {
	filename := li.Title + ".yaml"
	return ioutil.WriteFile(filename, []byte(li.Description), 0600)
}

func load_config(file string) (err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, globalConfig)

	// compute hashes for launch id:
	for i, item := range globalConfig.ItemList {
		item.Id = i
	}

	return
}

func kill_handler(w http.ResponseWriter, r *http.Request) {

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

func launch_handler(w http.ResponseWriter, r *http.Request) {

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

	err := load_config("data/items.yml")
	if err != nil {
		fmt.Println("failed to load config ", err)
	}

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handler)
	http.HandleFunc("/launch", launch_handler)
	http.HandleFunc("/kill", kill_handler)

	log.Fatal(http.ListenAndServe(":8181", nil))

}
