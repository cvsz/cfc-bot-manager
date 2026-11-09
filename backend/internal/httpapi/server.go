package httpapi

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
)

//go:embed web/*
var editorFiles embed.FS

type ReadinessCheck func(context.Context) error
type response struct { Status string `json:"status"`; Version string `json:"version,omitempty"` }

func NewHandler(version string, readiness ReadinessCheck, roomStyles ...http.Handler) http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,_ *http.Request){writeJSON(w,200,response{Status:"ok"})})
	mux.HandleFunc("GET /readyz",func(w http.ResponseWriter,r *http.Request){if readiness!=nil{if err:=readiness(r.Context());err!=nil{writeJSON(w,503,response{Status:"not_ready"});return}};writeJSON(w,200,response{Status:"ready"})})
	mux.HandleFunc("GET /api/v1/meta",func(w http.ResponseWriter,_ *http.Request){writeJSON(w,200,response{Status:"ok",Version:version})})
	mux.HandleFunc("GET /room-style-editor",func(w http.ResponseWriter,r *http.Request){http.Redirect(w,r,"/room-style-editor/",http.StatusTemporaryRedirect)})
	if sub,err:=fs.Sub(editorFiles,"web");err==nil{mux.Handle("GET /room-style-editor/",http.StripPrefix("/room-style-editor/",http.FileServer(http.FS(sub))))}
	if len(roomStyles)>0&&roomStyles[0]!=nil{mux.Handle("/api/v1/",roomStyles[0])}
	return securityHeaders(mux)
}
func securityHeaders(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
	w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Cache-Control","no-store");w.Header().Set("Referrer-Policy","no-referrer")
	w.Header().Set("Content-Security-Policy","default-src 'none'; style-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'")
	next.ServeHTTP(w,r)
})}
func writeJSON(w http.ResponseWriter,status int,v response){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v)}
