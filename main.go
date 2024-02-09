package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.New("").Funcs(template.FuncMap{"string": func(b []byte) string { return string(b) }}).ParseFiles("./static/index.html", "./static/create.html", "./static/view.html", "./static/edit.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	var directory string = "./Files/"
	title := strings.ReplaceAll(p.Title, " ", "")
	filename := title + ".txt"
	filePath := directory + filename

	return os.WriteFile(filePath, p.Body, 0600)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	fileDir := "./Files/" + filename
	body, err := os.ReadFile(fileDir)
	if err != nil {
		fmt.Println(err)
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	directory := "./Files"
	checkDir(directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var notes []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			notes = append(notes, strings.TrimSuffix(file.Name(), ".txt"))
		}
	}

	err = templates.ExecuteTemplate(w, "index.html", notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "create", nil)
	case http.MethodPost:
		title := r.FormValue("title")
		body := r.FormValue("body")

		if title == "" || body == "" {
			http.Error(w, "Title and body cannot be empty", http.StatusBadRequest)
			return
		}

		p := &Page{Title: title, Body: []byte(body)}
		err := p.save()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	fileServer := http.FileServer(http.Dir("./Static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/create", createHandler)
	serverURL := "http://localhost:3000"
	fmt.Printf("Server is started at: %s\n", serverURL)

	// Open the server URL in a browser
	openBrowser(serverURL)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func createDir(directory string) {
	err := os.Mkdir(directory, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
}

func checkDir(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Directory %s does not exist\n", directory)
		fmt.Printf("Creating directory:\n")
		createDir(directory)
		fmt.Printf("Directory Created\n")
	} else {
		fmt.Printf("Directory %s exists\n", directory)
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}