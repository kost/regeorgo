package regeorgo

import (
	"fmt"
	"net"
	"time"
	"net/http"
	"math/rand"
	"io/ioutil"
	"strings"
	"net/http/httputil"
)

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

type GeorgHandler struct {
	LogLevel int
	sm map[string]net.Conn
}

func (gh *GeorgHandler) InitHandler() {
	gh.sm = make(map[string]net.Conn)
}

func (gh *GeorgHandler) RegHandler(w http.ResponseWriter, r *http.Request) {
	if gh.LogLevel>1 {
		out, _ := httputil.DumpRequest(r, false)
		fmt.Printf("%q\n", out)
	}

	if r.Method=="GET" {
		fmt.Fprintln(w, "Georg says, 'All seems fine'")
		return
	}
	if r.Method=="POST" {
		cmd := strings.ToUpper(r.Header.Get("X-CMD"))
		// fmt.Printf("[cmd] %s\n", cmd)
		switch cmd {
			case "CONNECT":
				thost := r.Header.Get("X-TARGET")
				tport := r.Header.Get("X-PORT")
				hostport := fmt.Sprintf("%s:%s", thost, tport)

				//fmt.Printf("[connect] Content-Length: %d, %s\n", r.ContentLength, hostport)

				tcpAddr, _ := net.ResolveTCPAddr("tcp4", hostport)
				conn, err := net.DialTCP("tcp", nil, tcpAddr)
				if err != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("%v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				strkey := RandomString(32)
				gh.sm[strkey]=conn
				expiration := time.Now().Add(365 * 24 * time.Hour)
				cookie    :=  http.Cookie{Name: "socket",Value:strkey,Expires:expiration}
				http.SetCookie(w, &cookie)
				w.Header().Set("X-STATUS", "OK")

			case "DISCONNECT":
				// fmt.Printf("[disconnect] Content-Length: %d\n", r.ContentLength)

				cookie, err := r.Cookie("socket")
				if err != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("%v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				strkey := cookie.Value
				s, ok := gh.sm[strkey]
				if ok {
					s.Close()
					delete(gh.sm, strkey)
				} else {
					// fmt.Printf("[disconnect] key does not exists: %s\n", strkey)
				}
				c := &http.Cookie{
				    Name:     "socket",
				    Value:    "",
				    Path:     "/",
				    Expires: time.Unix(0, 0),
				}
				http.SetCookie(w, c)
				w.Header().Set("X-STATUS", "OK")

			case "READ":
				// fmt.Printf("[read] Content-Length: %d\n", r.ContentLength)

				cookie, err := r.Cookie("socket")
				if err != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("r.Cookie(): %v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				strkey := cookie.Value
				sr, sok := gh.sm[strkey]
				if !sok {
					// fmt.Printf("[read] key does not exists: %s\n", strkey)
					w.Header().Set("X-ERROR",fmt.Sprintf("Key does not exist: %s", strkey))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				w.Header().Set("X-STATUS", "OK")
				for {
					buf := make([]byte, 512)
					lenb, read_err := sr.Read(buf)
					if read_err != nil {
						//fmt.Printf("[read] error : %v\n", read_err)
						w.Header().Set("X-ERROR",fmt.Sprintf("[read] ioutil.ReadAll(): %v", read_err))
						w.Header().Set("X-STATUS", "FAIL")
						return
					}
					if lenb < 1 {
						// fmt.Printf("[read] break: %q\n", buf)
						break
					}
					_, write_err := w.Write(buf[:lenb])
					if write_err != nil {
						// fmt.Printf("[write] error (%d)\n", n)
						w.Header().Set("X-ERROR",fmt.Sprintf("[read] Write(): %v", write_err))
						w.Header().Set("X-STATUS", "FAIL")
						return
					}
					//sbuf := string(buf[:lenb])
					//fmt.Printf("[read] loop (%d, %d): %s\n", lenb, n, sbuf)
					if lenb < 512 {
						//fmt.Printf("[read] break (%d): %q\n", lenb, sbuf)
						break
					}
				}
				//fmt.Printf("[read] complete loop\n")

			case "FORWARD":
				// fmt.Printf("[forward] Content-Length: %d\n", r.ContentLength)

				cookie, err := r.Cookie("socket")
				if err != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("r.Cookie(): %v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				strkey := cookie.Value
				sf, fok := gh.sm[strkey]
				if !fok {
					// fmt.Printf("[forward] key does not exists: %s\n", strkey)
					w.Header().Set("X-ERROR",fmt.Sprintf("Key does not exist: %s", strkey))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}

				body, rerr := ioutil.ReadAll(r.Body)
				// fmt.Printf("[forward] got body: %s\n", body)
				if rerr != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("ioutil.ReadAll(): %v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				if closeErr := r.Body.Close(); closeErr != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("r.Body.Close(): %v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				_, write_err := sf.Write(body)
				if write_err != nil {
					w.Header().Set("X-ERROR",fmt.Sprintf("Write(): %v", err))
					w.Header().Set("X-STATUS", "FAIL")
					return
				}
				w.Header().Set("X-STATUS", "OK")
			default:
				// r.ContentLength
				// fmt.Printf("[error] unknown cmd: %s\n", cmd)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
		} // switch
	} // POST
} // func
