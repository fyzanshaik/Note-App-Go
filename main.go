package main

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Page struct {
	Title     string
	Body      []byte
	Timestamp string
}

var templates = template.Must(template.New("").Funcs(template.FuncMap{"string": func(b []byte) string { return string(b) }}).ParseFiles(
	"./static/index.html", "./static/create.html", "./static/view.html", "./static/edit.html", "./static/delete.html"))
var validPath = regexp.MustCompile(`^/(edit|save|view|delete)/([a-zA-Z0-9_/-]+)$`)

func (p *Page) save() error {
	var directory string = "./Files/"
	title := strings.ReplaceAll(p.Title, " ", "_")
	filename := title + ".md"
	filePath := directory + filename

	p.Body = []byte(fmt.Sprintf("# %s\n\n%s\n\n*Created on: %s*", p.Title, string(p.Body), p.Timestamp))

	return os.WriteFile(filePath, p.Body, 0600)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	fmt.Println("Path: ", r.URL.Path)
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
	filename := title + ".md"
	fileDir := "./Files/" + filename
	body, err := os.ReadFile(fileDir)
	if err != nil {
		return nil, err
	}
	title = strings.ReplaceAll(title, "_", " ")
	body, timeStamp := bodyHandler(body)
	return &Page{Title: title, Body: body, Timestamp: string(timeStamp)}, nil
}

func bodyHandler(dataStream []uint8) ([]uint8, []uint8) {
	dataStr := string(dataStream)

	firstAsteriskPos := strings.IndexByte(dataStr, '*')
	lastAsteriskPos := strings.LastIndexByte(dataStr, '*')

	if firstAsteriskPos == -1 || lastAsteriskPos == -1 || lastAsteriskPos <= firstAsteriskPos {
		return []uint8{}, []uint8{}
	}

	newlinePos := strings.IndexByte(dataStr, '\n')
	if newlinePos == -1 || firstAsteriskPos <= newlinePos {
		return []uint8{}, []uint8{}
	}

	startPos := newlinePos + 1
	bodyStr := strings.TrimSpace(dataStr[startPos:firstAsteriskPos])

	timestampStr := strings.TrimSpace(dataStr[firstAsteriskPos+1 : lastAsteriskPos])

	bodySlice := []uint8(bodyStr)
	timestampSlice := []uint8(timestampStr)

	return bodySlice, timestampSlice
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	// title = strings.ReplaceAll(title,"_"," ");
	fmt.Println("Title: ", title)
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
	p := &Page{Title: title, Body: []byte(body), Timestamp: time.Now().Format(time.RFC1123)}
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

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	filename := title + ".md"
	fileDir := "./Files/" + filename
	err = os.Remove(fileDir)
	if err != nil {
		http.Error(w, "Unable to delete the entry", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	directory := "./Files"
	checkDir(directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	type EntryInfo struct {
		Title     string
		Path      string
		Timestamp string
	}

	var entries []EntryInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			realPathTitle := strings.TrimSuffix(file.Name(), ".md")
			title := strings.ReplaceAll(realPathTitle, "_", " ")
			timestamp := readTimestampFromFile(directory + "/" + file.Name())
			fmt.Printf("File: %s, Timestamp: %s\n\n", file.Name(), timestamp)
			entries = append(entries, EntryInfo{
				Title:     title,
				Path:      realPathTitle,
				Timestamp: timestamp,
			})
		}
	}

	const layout = "Mon, 02 Jan 2006 15:04:05 MST"

	sort.Slice(entries, func(i, j int) bool {
		timestampI, errI := time.Parse(layout, entries[i].Timestamp)
		timestampJ, errJ := time.Parse(layout, entries[j].Timestamp)

		if errI != nil || errJ != nil {
			return false
		}

		return timestampI.After(timestampJ)
	})

	err = templates.ExecuteTemplate(w, "index.html", entries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func readTimestampFromFile(filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filepath, err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "*Created on:") {
			timestampStr := strings.TrimPrefix(line, "*Created on:")
			timestampStr = strings.TrimSpace(timestampStr)
			timestampStr = strings.TrimSuffix(timestampStr, "*")
			fmt.Printf("Raw timestamp string: %s\n", timestampStr)

			timestamp, err := time.Parse(time.RFC1123, timestampStr)
			if err == nil {
				formattedTime := timestamp.Format("2006-01-02 15:04:05")
				fmt.Printf("Parsed and formatted timestamp: %s\n", formattedTime)
				return formattedTime
			} else {
				fmt.Printf("Error parsing timestamp: %v\n", err)
				return timestampStr
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filepath, err)
	}

	fmt.Printf("No timestamp found in file %s\n", filepath)
	return ""
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

		p := &Page{Title: title, Body: []byte(body), Timestamp: time.Now().Format(time.RFC1123)}
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
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/create", createHandler)

	serverURL := "http://localhost:3000"
	fmt.Printf("Server is started at: %s\n", serverURL)

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
