package main

import (
	"./session"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

var ch map[string](chan [2]string)

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

type Erro struct {
	Err   string
	Uname string
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if sess != nil && err == nil {
		http.Redirect(w, r, "/chatengine", http.StatusFound)
	}
	if r.Method != "POST" {
		renderTemplate(w, "login.html", &Erro{Err: ""})
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
		renderTemplate(w, "login.html", &Erro{Err: "Couldnt log in"})
		return
	}
	log.Printf("Username = %v Password = %v", uname, pass)
	if p, ok := users[uname]; !(ok && p == pass) {
		renderTemplate(w, "login.html", &Erro{Err: "Wrong username or password"})
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
		close(ch[(*sess)["Username"]])
		delete(ch, (*sess)["Username"])
		quit[(*sess)["Username"]] <- true
		hd.SessionDestroy(&w, r)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	ch[(*sess)["Username"]] = make(chan [2]string, 2)
	quit[(*sess)["Username"]] = make(chan bool)
	log.Printf("Session is %#v\n", *sess)
	log.Printf("Channel is %#v\n", ch)
	renderTemplate(w, "chat.html", &Erro{Uname: (*sess)["Username"]})
}

func listenMsg(w http.ResponseWriter, r *http.Request, uname string) {
	f, ok := w.(http.Flusher)
	if !ok {
		renderTemplate(w, "error.html", nil)
		return
	}
	for {
		select {
		case msg := <-ch[uname]:
			log.Printf("Got msg %v by %s", msg, uname)
			msg1 := "<div class='chatmsg'><b>" + msg[0] + " says</b> " + msg[1] + "<br/></div>"
			fmt.Fprintf(w, "data: %s\n\n", msg1)
			f.Flush()
		case <-quit[uname]:
			log.Printf("Closing conn for %s", uname)
			close(quit[uname])
			delete(quit, uname)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
			// default:
			// 	log.Printf("Nothing")
		}
	}
}

func chatlistenHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	log.Printf("Inside chatHandler of %s", (*sess)["Username"])

	listenMsg(w, r, (*sess)["Username"])

}

func sendto(uname string, msg1 [2]string) {
	if _, ok := ch[uname]; ok {
		log.Printf("msg %#v sent\n", msg1)
		ch[uname] <- msg1
	} else {
		//write to db
	}
}

func sendmsgHandler(w http.ResponseWriter, r *http.Request) {
	hd := session.GetSessionHandler()
	sess, err := hd.SessionStart(&w, r)
	log.Printf(r.Method)
	if err != nil || sess == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if r.Method != "POST" {
		renderTemplate(w, "error.html", nil)
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("error '%s' parsing form for %#v\n", err, r)
		return
	}
	uname := r.Form.Get("uname")
	msg := r.Form.Get("msg")
	msg1 := [2]string{(*sess)["Username"], msg}
	log.Printf("msg '%s' is %#v\n", uname, msg1)
	log.Printf("Channel is %#v\n", ch)
	log.Printf("Session is %#v\n", *sess)
	go sendto(uname, msg1)
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	renderTemplate(w, "chat.js", nil)
}

func main() {
	ch = make(map[string](chan [2]string))
	quit = make(map[string](chan bool))
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/chatengine", chatHandler)
	http.HandleFunc("/chatlisten", chatlistenHandler)
	http.HandleFunc("/sendmsg", sendmsgHandler)
	http.HandleFunc("/chat.js", jsHandler)
	//http.HandleFunc("/save/", saveHandler)
	http.ListenAndServe(":8080", nil)
}
