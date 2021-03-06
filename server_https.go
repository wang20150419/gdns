package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fangdingjun/go-log"
	"github.com/miekg/dns"
)

const dnsMsgType = "application/dns-message"

func (srv *server) handleHTTPReq(w http.ResponseWriter, r *http.Request) {
	ctype := r.Header.Get("content-type")
	if !strings.HasPrefix(ctype, dnsMsgType) {
		log.Errorf("request type %s, require %s", ctype, dnsMsgType)
		http.Error(w, "dns message is required", http.StatusBadRequest)
		return
	}

	if r.ContentLength < 10 {
		log.Errorf("message is too small, %v", r.ContentLength)
		http.Error(w, "message is too small", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln("read request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msg := new(dns.Msg)
	if err := msg.Unpack(data); err != nil {
		log.Errorln("parse dns message", err)
		return
	}
	m, err := getResponseFromUpstream(msg, srv.upstreams)
	if err != nil {
		log.Debugln("query", msg.Question[0].String(), "timeout")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	for _, a := range m.Answer {
		log.Debugln("result", a.String())
	}
	w.Header().Set("content-type", dnsMsgType)
	d, _ := m.Pack()
	w.Write(d)
}

func (srv *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != srv.addr.Path {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	srv.handleHTTPReq(w, r)
}
