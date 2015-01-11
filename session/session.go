package session

import (
	"./uniuri"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"time"
)

type sessionData map[string]string

type SessionHandler struct {
	CookieName string
	CookiePath string
	httponly   bool
}

var shd = SessionHandler{
	CookieName: "chat-msg",
	CookiePath: "/",
	httponly:   true,
}

var sessionDB map[string]*sessionData = make(map[string]*sessionData)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (h *SessionHandler) SessionDestroy(wr *http.ResponseWriter, req *http.Request) error {
	cookie, err := req.Cookie(h.CookieName)
	if err != nil || cookie.Value == "" {
		return err
	}
	if _, ok := sessionDB[cookie.Value]; ok {
		delete(sessionDB, cookie.Value)
	}
	var newcookie http.Cookie
	newcookie.Name = h.CookieName
	newcookie.Value = ""
	newcookie.Path = "/"
	newcookie.MaxAge = -10
	newcookie.Expires = time.Unix(0, 0).UTC()
	http.SetCookie(*wr, &newcookie)
	return nil
}

func (h *SessionHandler) SessionStart(wr *http.ResponseWriter, req *http.Request) (*sessionData, error) {
	cookie, err := req.Cookie(h.CookieName)
	if err != nil || cookie.Value == "" {
		return nil, err
	}
	if _, ok := sessionDB[cookie.Value]; ok {
		return sessionDB[cookie.Value], nil
	}
	h.SessionDestroy(wr, req)
	return nil, nil
}

func (h *SessionHandler) SessionCreate(wr *http.ResponseWriter, req *http.Request, username string) *sessionData {
	cookie, err := req.Cookie(h.CookieName)
	if err != nil || cookie.Value == "" {
		var cookie http.Cookie
		cookie.Name = h.CookieName
		cookie.Value = GetMD5Hash(uniuri.New() + username)
		cookie.Path = h.CookiePath
		cookie.HttpOnly = h.httponly
		cookie.Secure = false
		http.SetCookie(*wr, &cookie)
		sessionDB[cookie.Value] = &sessionData{"Username": username}
		return sessionDB[cookie.Value]
	}
	if _, ok := sessionDB[cookie.Value]; ok {
		return sessionDB[cookie.Value]
	}
	h.SessionDestroy(wr, req)
	return nil
}

func GetSessionHandler() *SessionHandler {
	return &shd
}
