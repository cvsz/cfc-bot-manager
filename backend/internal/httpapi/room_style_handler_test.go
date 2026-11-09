package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
)

type memoryStyleStore struct{mu sync.RWMutex; defaultStyle messagestyle.Style; rooms map[string]messagestyle.Style}
func newMemoryStyleStore()*memoryStyleStore{return &memoryStyleStore{defaultStyle:messagestyle.DefaultStyle(),rooms:make(map[string]messagestyle.Style)}}
func(s *memoryStyleStore)GetDefault(context.Context)(messagestyle.Style,error){s.mu.RLock();defer s.mu.RUnlock();return s.defaultStyle,nil}
func(s *memoryStyleStore)SaveDefault(_ context.Context,v messagestyle.Style)error{s.mu.Lock();defer s.mu.Unlock();s.defaultStyle=v;return nil}
func(s *memoryStyleStore)GetRoom(_ context.Context,id string)(messagestyle.Style,bool,error){s.mu.RLock();defer s.mu.RUnlock();v,ok:=s.rooms[id];return v,ok,nil}
func(s *memoryStyleStore)SaveRoom(_ context.Context,id string,v messagestyle.Style)error{s.mu.Lock();defer s.mu.Unlock();s.rooms[id]=v;return nil}
func(s *memoryStyleStore)DeleteRoom(_ context.Context,id string)error{s.mu.Lock();defer s.mu.Unlock();delete(s.rooms,id);return nil}
func styleRequest(h http.Handler,method,path,body string,auth bool)*httptest.ResponseRecorder{
	r:=httptest.NewRequest(method,path,strings.NewReader(body));if auth{r.Header.Set("Authorization","Bearer 0123456789abcdef0123456789abcdef")}
	w:=httptest.NewRecorder();h.ServeHTTP(w,r);return w
}
func TestRoomStyleRequiresBearerToken(t *testing.T){
	h:=NewRoomStyleHandler(newMemoryStyleStore(),"0123456789abcdef0123456789abcdef")
	if got:=styleRequest(h,"GET","/api/v1/message-style/default","",false).Code;got!=401{t.Fatalf("status=%d",got)}
}
func TestRoomStyleSaveLoadAndReset(t *testing.T){
	h:=NewRoomStyleHandler(newMemoryStyleStore(),"0123456789abcdef0123456789abcdef")
	body:=`{"fontFamily":"Tahoma","fontSize":18,"color":"#aabbcc","bold":true,"italic":false,"underline":true}`
	saved:=styleRequest(h,"PUT","/api/v1/rooms/room-42/message-style",body,true)
	if saved.Code!=200||!strings.Contains(saved.Body.String(),`"source":"room"`)||!strings.Contains(saved.Body.String(),`"#AABBCC"`){t.Fatalf("save: %d %s",saved.Code,saved.Body.String())}
	loaded:=styleRequest(h,"GET","/api/v1/rooms/room-42/message-style","",true)
	if loaded.Code!=200||!strings.Contains(loaded.Body.String(),`"fontFamily":"Tahoma"`){t.Fatalf("load: %d %s",loaded.Code,loaded.Body.String())}
	reset:=styleRequest(h,"DELETE","/api/v1/rooms/room-42/message-style","",true)
	if reset.Code!=200||!strings.Contains(reset.Body.String(),`"source":"global_default"`){t.Fatalf("reset: %d %s",reset.Code,reset.Body.String())}
}
func TestRoomStyleValidationAndFallback(t *testing.T){
	h:=NewRoomStyleHandler(newMemoryStyleStore(),"0123456789abcdef0123456789abcdef")
	fallback:=styleRequest(h,"GET","/api/v1/rooms/room-1/message-style","",true)
	if fallback.Code!=200||!strings.Contains(fallback.Body.String(),`"source":"global_default"`){t.Fatalf("fallback: %d %s",fallback.Code,fallback.Body.String())}
	invalid:=styleRequest(h,"PUT","/api/v1/rooms/room-1/message-style",`{"fontFamily":"Comic Sans","fontSize":999,"color":"red"}`,true)
	if invalid.Code!=422{t.Fatalf("invalid style status=%d",invalid.Code)}
	unknown:=styleRequest(h,"PUT","/api/v1/rooms/room-1/message-style",`{"fontFamily":"Arial","fontSize":14,"color":"#FFFFFF","extra":true}`,true)
	if unknown.Code!=400{t.Fatalf("unknown field status=%d",unknown.Code)}
}
func TestRoomStyleRejectsOversizedBody(t *testing.T){
	h:=NewRoomStyleHandler(newMemoryStyleStore(),"0123456789abcdef0123456789abcdef")
	body:=`{"fontFamily":"Arial","fontSize":14,"color":"#FFFFFF","extra":"`+strings.Repeat("x",5000)+`"}`
	got:=styleRequest(h,"PUT","/api/v1/message-style/default",body,true).Code
	if got!=400&&got!=413{t.Fatalf("oversized body status=%d",got)}
}
