package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
)

var roomIDPattern=regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$")
type roomStyleResponse struct { Style messagestyle.Style `json:"style"`; Source string `json:"source"` }
type apiError struct { Error string `json:"error"` }

func NewRoomStyleHandler(store messagestyle.Store,token string)http.Handler{
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /api/v1/message-style/default",func(w http.ResponseWriter,r *http.Request){
		v,err:=store.GetDefault(r.Context());if err!=nil{writeAPIError(w,500,"internal_error");return};writeAPIJSON(w,200,roomStyleResponse{v,"global_default"})
	})
	mux.HandleFunc("PUT /api/v1/message-style/default",func(w http.ResponseWriter,r *http.Request){
		v,ok:=decodeStyle(w,r);if !ok{return};if err:=store.SaveDefault(r.Context(),v);err!=nil{writeAPIError(w,500,"internal_error");return};writeAPIJSON(w,200,roomStyleResponse{v,"global_default"})
	})
	mux.HandleFunc("GET /api/v1/rooms/{roomID}/message-style",func(w http.ResponseWriter,r *http.Request){
		id:=r.PathValue("roomID");if !validRoomID(id){writeAPIError(w,400,"invalid_room_id");return}
		v,found,err:=store.GetRoom(r.Context(),id);if err!=nil{writeAPIError(w,500,"internal_error");return};source:="room"
		if !found{v,err=store.GetDefault(r.Context());source="global_default"};if err!=nil{writeAPIError(w,500,"internal_error");return};writeAPIJSON(w,200,roomStyleResponse{v,source})
	})
	mux.HandleFunc("PUT /api/v1/rooms/{roomID}/message-style",func(w http.ResponseWriter,r *http.Request){
		id:=r.PathValue("roomID");if !validRoomID(id){writeAPIError(w,400,"invalid_room_id");return};v,ok:=decodeStyle(w,r);if !ok{return}
		if err:=store.SaveRoom(r.Context(),id,v);err!=nil{writeAPIError(w,500,"internal_error");return};writeAPIJSON(w,200,roomStyleResponse{v,"room"})
	})
	mux.HandleFunc("DELETE /api/v1/rooms/{roomID}/message-style",func(w http.ResponseWriter,r *http.Request){
		id:=r.PathValue("roomID");if !validRoomID(id){writeAPIError(w,400,"invalid_room_id");return}
		if err:=store.DeleteRoom(r.Context(),id);err!=nil{writeAPIError(w,500,"internal_error");return};v,err:=store.GetDefault(r.Context());if err!=nil{writeAPIError(w,500,"internal_error");return};writeAPIJSON(w,200,roomStyleResponse{v,"global_default"})
	})
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if !authorized(r,token){w.Header().Set("WWW-Authenticate","Bearer");writeAPIError(w,401,"unauthorized");return};mux.ServeHTTP(w,r)
	})
}
func authorized(r *http.Request,token string)bool{
	if token==""{return false};h:=r.Header.Get("Authorization");if !strings.HasPrefix(h,"Bearer "){return false}
	got:=strings.TrimPrefix(h,"Bearer ");if got==""||strings.ContainsAny(got," \t\r\n"){return false}
	return subtle.ConstantTimeCompare([]byte(got),[]byte(token))==1
}
func validRoomID(v string)bool{return roomIDPattern.MatchString(v)}
func decodeStyle(w http.ResponseWriter,r *http.Request)(messagestyle.Style,bool){
	r.Body=http.MaxBytesReader(w,r.Body,4096);d:=json.NewDecoder(r.Body);d.DisallowUnknownFields();var v messagestyle.Style
	if err:=d.Decode(&v);err!=nil{var tooLarge *http.MaxBytesError;if errors.As(err,&tooLarge){writeAPIError(w,413,"request_too_large")}else{writeAPIError(w,400,"invalid_json")};return messagestyle.Style{},false}
	if err:=d.Decode(&struct{}{});err!=io.EOF{writeAPIError(w,400,"invalid_json");return messagestyle.Style{},false}
	normalized,err:=v.Normalize();if err!=nil{writeAPIError(w,422,"invalid_style");return messagestyle.Style{},false};return normalized,true
}
func writeAPIJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v)}
func writeAPIError(w http.ResponseWriter,status int,code string){writeAPIJSON(w,status,apiError{code})}
