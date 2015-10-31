package main

import (
	"net/http"

	"golang.org/x/net/context"
)

func NewContext(parent context.Context, s *Server, req *http.Request) *Context {
	return &Context{Context: parent, Server: s, Request: req}
}

type Context struct {
	context.Context

	Request *http.Request
	Server  *Server
}
