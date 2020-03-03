package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

type Config struct {
	LauncherItemList []LauncherItem `yaml:"items"`
}

type LauncherItem struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func save(li *LauncherItem) error {
	filename := li.Title + ".yaml"
	return ioutil.WriteFile(filename, []byte(li.Description), 0600)
}

func load_config(file string) (conf *Config, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	conf = &Config{}
	err = yaml.Unmarshal(data, conf)
	return
}

func handler(w http.ResponseWriter, r *http.Request) {

	cfg, err := load_config("data/items.yml")
	if err != nil {
		fmt.Println("failed to load config ", err)
	}

	for _, s := range cfg.LauncherItemList {
		fmt.Println("\n++++ Title = ", s.Title)
	}

	// tmpl, err := template.ParseFiles("templates/view.html")

	// tmpl.Execute(w, cfg.LauncherItemList)

	fmt.Fprint(w, "Hello world!", cfg)
}

func main() {

	fmt.Printf("GIL Launcher\n")

	http.HandleFunc("/", handler)

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
