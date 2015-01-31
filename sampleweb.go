package main

import (
	"./session"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var ch map[string](chan [3]string)

var quit map[string](chan bool)

var users = map[string](string){
	"asd": "123",
	"qwe": "123",
	"zxc": "123",
}

type Page struct {
	Title string
	Body  []byte
}

type Tmpl struct {
	Err   string
	Uname string
	Body  string
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	t, _ := template.ParseFiles(tmpl)
	err := t.Execute(w, p)
	if err != nil {
		log.Printf("error '%s'\n", err)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, _ := loadPage(title)
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit.html", p)
}

func isloggedinHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("isloggedin %#v\n", r.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	log.Printf("is logged in\n")
	if r.Method != "POST" {
		fmt.Fprintf(w, "0")
		return
	}
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if sess != nil && err == nil {
		fmt.Fprintf(w, "1")
		return
	}
	fmt.Fprintf(w, "0")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if sess != nil && err == nil {
		http.Redirect(w, r, "/chatengine", http.StatusFound)
	}
	if r.Method != "POST" {
		renderTemplate(w, "login1.html", &Tmpl{Err: ""})
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("error '%s' parsing form for %#v\n", err, r)
		return
	}
	uname := r.Form.Get("username")
	pass := r.Form.Get("password")
	if _, exists := ch[uname]; exists {
		renderTemplate(w, "login1.html", &Tmpl{Err: "Couldnt log in"})
		return
	}
	log.Printf("Username = %v Password = %v", uname, pass)
	if p, ok := users[uname]; !(ok && p == pass) {
		renderTemplate(w, "login1.html", &Tmpl{Err: "Wrong username or password"})
		return
	}
	sess = hd.SessionCreate(&w, r, uname)
	if sess != nil {
		http.Redirect(w, r, "/chatengine", http.StatusFound)
		return
	}
	renderTemplate(w, "error.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if sess != nil && err == nil {
		if _, exists := ch[(*sess)["Username"]]; exists {
			close(ch[(*sess)["Username"]])
			delete(ch, (*sess)["Username"])
		}
		if _, exists := quit[(*sess)["Username"]]; exists {
			quit[(*sess)["Username"]] <- true
		}
		hd.SessionDestroy(&w, r)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	// hd := session.GetSessionHandler()
	// sess, err := hd.SessionStart(&w, r)
	// if err != nil || sess == nil {
	// 	http.Redirect(w, r, "/login", http.StatusFound)
	// 	return
	// }
	// ch[(*sess)["Username"]] = make(chan [3]string, 2)
	// quit[(*sess)["Username"]] = make(chan bool)
	// log.Printf("Session is %#v\n", *sess)
	// log.Printf("Channel is %#v\n", ch)
	renderTemplate(w, "chatnew.html", &Tmpl{ /*Uname: (*sess)["Username"]*/ })
}

func listenMsg(w *http.ResponseWriter, r *http.Request, uname string, f *http.Flusher) {
	(*w).Header().Set("Content-Type", "text/event-stream;charset=UTF-8")
	(*w).Header().Set("Cache-Control", "no-cache")
	(*w).Header().Set("Connection", "keep-alive")
	fmt.Fprintf(*w, "data: Ready\n\n")
	(*f).Flush()
	log.Printf("Going to for of %s", uname)
	for {
		select {
		case msg := <-ch[uname]:
			log.Printf("Got msg %v by %s", msg, uname)
			switch msg[2] {
			case "msg":
				fmt.Fprintf((*w), "event: msg\ndata:{%s}\n\n", msg[1])
			case "canvas":
				fmt.Fprintf((*w), "event: canvas\ndata:{%s}\n\n", msg[1])
			case "file":
				fmt.Fprintf((*w), "event: file\ndata:{%s}\n\n", msg[1])
			}
			(*f).Flush()
		case <-quit[uname]:
			log.Printf("Closing conn for %s", uname)
			close(quit[uname])
			delete(quit, uname)
			// http.Redirect((*w), r, "/login", http.StatusFound)
			return
			// default:
			// 	log.Printf("Nothing")
		}
	}
}

func chatlistenHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	log.Printf("Inside chatHandler of %s", (*sess)["Username"])
	f, ok := w.(http.Flusher)
	if !ok {
		renderTemplate(w, "error.html", nil)
		return
	}
	listenMsg(&w, r, (*sess)["Username"], &f)
}

func sendto(uname string, msg1 [3]string) {
	if _, ok := ch[uname]; ok {
		ch[uname] <- msg1
		log.Printf("msg %#v sent\n", msg1)
	} else {
		//write to db
	}
}

func sendmsgHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	log.Printf(r.Method)
	if err != nil || sess == nil {
		fmt.Fprintf(w, "0")
		return
	}
	if r.Method != "POST" {
		fmt.Fprintf(w, "0")
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("error '%s' parsing form for %#v\n", err, r)
		fmt.Fprintf(w, "0")
		return
	}
	uname := r.Form.Get("uname")
	msg := r.Form.Get("msg")
	str := "\"uname\": \"" + (*sess)["Username"] + "\", \"msg\": \"" + msg + "\""
	msg1 := [3]string{(*sess)["Username"], str, "msg"}
	log.Printf("msg '%s' is %#v\n", uname, msg1)
	log.Printf("Channel is %#v\n", ch)
	log.Printf("Session is %#v\n", *sess)
	go sendto(uname, msg1)
	fmt.Fprintf(w, "1")
}

func sendcanvasHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	log.Printf("Got canvas")
	if err != nil || sess == nil {
		fmt.Fprintf(w, "0")
		return
	}
	if r.Method != "POST" {
		fmt.Fprintf(w, "0")
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("error '%s' parsing form for %#v\n", err, r)
		fmt.Fprintf(w, "0")
		return
	}
	bck := r.Form.Get("bck")
	uname := r.Form.Get("uname")
	img := r.Form.Get("img")
	height := r.Form.Get("height")
	width := r.Form.Get("width")
	str := "\"bck\": \"" + bck + "\", \"img\": \"" + img + "\", \"nheight\": \"" + height + "\", \"nwidth\": \"" + width + "\", \"uname\": \"" + (*sess)["Username"] + "\""
	msg1 := [3]string{(*sess)["Username"], str, "canvas"}
	log.Printf("msg '%s' is %#v\n", uname, msg1)
	log.Printf("Channel is %#v\n", ch)
	log.Printf("Session is %#v\n", *sess)
	go sendto(uname, msg1)
	fmt.Fprintf(w, "1")
}

func sendfileHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	log.Printf(r.Method)
	if err != nil || sess == nil {
		fmt.Fprintf(w, "0")
		return
	}
	if r.Method != "POST" {
		fmt.Fprintf(w, "0")
		return
	}

	uname := r.FormValue("uname")
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}
	defer file.Close()

	filename := GetMD5Hash(strings.Replace(time.Now().Format("20060102150405")+(*sess)["Username"]+header.Filename, " ", "_", -1))
	filename = strings.Replace(filename, "/", "", -1)
	log.Println(uname)
	err = os.MkdirAll("tmp/"+uname, 0777)
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}
	out, err := os.Create("tmp/" + uname + "/" + filename)
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintf(w, "0")
		return
	}

	str := "\"filename\": \"" + filename + "\", \"file\": \"" + header.Filename + "\", \"uname\": \"" + (*sess)["Username"] + "\""
	msg1 := [3]string{(*sess)["Username"], str, "file"}
	go sendto(uname, msg1)
	fmt.Fprintf(w, "1")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	log.Printf(r.Method)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	filename := r.URL.Path[len("/download/"):]
	fi, err := os.Stat("tmp/" + (*sess)["Username"] + "/" + filename)
	if os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s", filename)
		return
	}

	log.Println(filename[17:])
	w.Header().Set("Content-Disposition", "attachment; filename="+filename[17:])
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(fi.Size()), 10))
	f, err := os.Open("tmp/" + (*sess)["Username"] + "/" + filename)
	if err != nil {
		fmt.Printf("no such file or directory: %s", filename)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	http.ServeFile(w, r, r.URL.Path[1:])
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeFile(w, r, r.URL.Path[1:])
}

func imgHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func exitHandler(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}
func colorHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderTemplate(w, "color.html", &Tmpl{Uname: (*sess)["Username"]})
}

func loginmeHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if sess != nil && err == nil {
		fmt.Fprintf(w, "1")
		return
	}
	if r.Method != "POST" {
		fmt.Fprintf(w, "0")
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("error '%s' parsing form for %#v\n", err, r)
		fmt.Fprintf(w, "0")
		return
	}
	uname := r.Form.Get("username")
	pass := r.Form.Get("password")
	if _, exists := ch[uname]; exists {
		fmt.Fprintf(w, "0")
		return
	}
	log.Printf("Username = %v Password = %v", uname, pass)
	if p, ok := users[uname]; !(ok && p == pass) {
		fmt.Fprintf(w, "0")
		return
	}
	sess = hd.SessionCreate(&w, r, uname)
	if sess != nil {
		ch[(*sess)["Username"]] = make(chan [3]string, 2)
		quit[(*sess)["Username"]] = make(chan bool)
		fmt.Fprintf(w, "1")
		return
	}
	fmt.Fprintf(w, "0")
}

func main() {
	ch = make(map[string](chan [3]string))
	quit = make(map[string](chan bool))
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/isloggedin", isloggedinHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", loginmeHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/chatengine", chatHandler)
	http.HandleFunc("/chatlisten", chatlistenHandler)
	http.HandleFunc("/sendmsg", sendmsgHandler)
	http.HandleFunc("/sendfile", sendfileHandler)
	http.HandleFunc("/sendcanvas", sendcanvasHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/jquery.js", jsHandler)
	http.HandleFunc("/jquerymobile/jquery.mobile-1.4.5.min.css", cssHandler)
	http.HandleFunc("/jquerymobile/jquery.mobile-1.4.5.min.js", jsHandler)
	http.HandleFunc("/chat.js", jsHandler)
	http.HandleFunc("/drawing.js", jsHandler)
	http.HandleFunc("/colormap.gif", imgHandler)
	http.HandleFunc("/selectedcolor.gif", imgHandler)
	http.HandleFunc("/color", colorHandler)
	http.HandleFunc("/exit", exitHandler)
	http.ListenAndServe(":8080", nil)
}
