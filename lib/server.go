package cks

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Server struct {
	parent *Cks
}

type Resp struct {
	Code int
	Msg  string
	Data interface{}
}

func NewServer(parent *Cks) *Server {

	s := new(Server)
	s.parent = parent

	return s
}

func (s *Server) Start() {

	server := http.NewServeMux()

	// server.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {

	// 	fmt.Fprintln(res, "<html><body>Hello world!</body></html>")

	// })
	server.Handle("/", http.FileServer(http.Dir("./web")))

	server.HandleFunc("/getstudent", func(w http.ResponseWriter, req *http.Request) {

		sid := req.FormValue("sid")
		data := s.parent.RedisDB.GetStudent(sid)

		s.json(w, req, data)

	})

	server.HandleFunc("/checkin", func(w http.ResponseWriter, req *http.Request) {

		data := s.parent.RedisDB.CheckIn(req)
		s.json(w, req, data)

	})

	server.HandleFunc("/checkinfo", func(w http.ResponseWriter, req *http.Request) {

		class := req.FormValue("class")
		data := s.parent.RedisDB.Check(class)

		s.json(w, req, data)

	})

	server.HandleFunc("/checkallinfo", func(w http.ResponseWriter, req *http.Request) {

		data := s.parent.RedisDB.Check("All")

		s.json(w, req, data)

	})

	// go http.ListenAndServe("0.0.0.0:2333", server)
	go http.ListenAndServeTLS("0.0.0.0:2333", "cert.pem", "privkey.pem", server)
	s.parent.logger.WithFields(logrus.Fields{
		"scope": "server/Start",
	}).Info("Listing 0.0.0.0:2333")

}

func (s *Server) json(w http.ResponseWriter, req *http.Request, data interface{}) {

	buf, err := json.Marshal(data)

	if err != nil {
		s.parent.logger.WithFields(logrus.Fields{
			"scope": "server/json",
		}).Fatal(err)
	}

	w.Write(buf)

}
