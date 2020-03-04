package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ItemList []Item `yaml:"items"`
}

type Item struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

func save(li *Item) error {
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

func launch_app(app string, args ...string) {

	mCmd := exec.Command(app)

	mCmdIn, _ := mCmd.StdinPipe()
	mCmdOut, _ := mCmd.StdoutPipe()

	mCmd.Start()

	mCmdIn.Close()
	outputBytes, _ := ioutil.ReadAll(mCmdOut)
	mCmd.Wait()

	_ = outputBytes
	_ = mCmdIn
	_ = mCmdOut

}

func handler(w http.ResponseWriter, r *http.Request) {

	cfg, err := load_config("data/items.yml")
	if err != nil {
		fmt.Println("failed to load config ", err)
	}

	for _, s := range cfg.ItemList {
		fmt.Println("\n++++ Title = ", s.Title)
	}

	tmpl, err := template.ParseFiles("templates/view.html")
	tmpl.Execute(w, cfg)

	// fmt.Fprint(w, "Config %S", reflect.TypeOf(cfg).String())
}

func main() {

	fmt.Printf("GIL Launcher\n")

	launch_app("firefox", "--kiosk", "http://localhost:8080")

	http.HandleFunc("/", handler)

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
