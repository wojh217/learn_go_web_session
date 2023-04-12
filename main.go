package main

import (
	"fmt"
	"github.com/wojh217/learn_go_web_session/session"
	_ "github.com/wojh217/learn_go_web_session/session/providers/memory"
	"net/http"
	"time"
)

var globalSessions *session.Manager

func init() {
	// 使用内存存储session
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)

	go globalSessions.GC()
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r) // 从cookie中获取value，也就是sessionid，没有则新建
	r.ParseForm()

	if r.Method == "GET" {
		username := sess.Get("username").(string)
		w.Write([]byte(fmt.Sprintf("hello, %s", username)))
	} else {
		sess.Set("username", r.Form["username"])
		w.Write([]byte(fmt.Sprintf("receive %s", r.Form["username"])))
	}
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	sess := globalSessions.SessionStart(w, r)
	createtime := sess.Get("createtime")
	if createtime == nil {
		sess.Set("craetetime", time.Now().Unix())
	} else if createtime.(int64)+360 < time.Now().Unix() {
		globalSessions.SessionDestroy(w, r)
		sess = globalSessions.SessionStart(w, r)
	}

	ct := sess.Get("countnum")
	if ct == nil {
		sess.Set("countnum", 1)
	} else {
		sess.Set("countnum", ct.(int)+1)
	}

	w.Write([]byte(fmt.Sprintf("cur num: %v\n", sess.Get("countnum"))))

}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/count", countHandler)

	http.ListenAndServe(":8099", nil)
}
