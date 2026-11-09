package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/config"
	"github.com/cvsz/cfc-bot-manager/backend/internal/httpapi"
	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var version="dev"
func main(){
	cfg,err:=config.Load();if err!=nil{slog.Error("invalid configuration","error",err);os.Exit(2)}
	logger:=slog.New(slog.NewJSONHandler(os.Stderr,&slog.HandlerOptions{Level:slog.LevelInfo}));slog.SetDefault(logger)
	var db *sql.DB;var readiness httpapi.ReadinessCheck;var styleHandler http.Handler
	if cfg.DatabaseURL!=""{
		db,err=sql.Open("pgx",cfg.DatabaseURL);if err!=nil{logger.Error("database initialization failed","error",err);os.Exit(1)};defer db.Close()
		db.SetMaxOpenConns(10);db.SetMaxIdleConns(5);db.SetConnMaxLifetime(30*time.Minute)
		ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);err=db.PingContext(ctx);cancel()
		if err!=nil{logger.Error("database unavailable");os.Exit(1)}
		readiness=func(ctx context.Context)error{return db.PingContext(ctx)}
		styleHandler=httpapi.NewRoomStyleHandler(messagestyle.NewPostgresStore(db),cfg.RoomStyleToken)
	}
	server:=NewServer(cfg,version,readiness,styleHandler)
	serveErr:=make(chan error,1)
	go func(){logger.Info("HTTP server starting","address",cfg.HTTPAddr,"environment",cfg.Environment,"version",version);if err:=server.ListenAndServe();err!=nil&&!errors.Is(err,http.ErrServerClosed){serveErr<-err;return};serveErr<-nil}()
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer stop()
	select{
	case <-ctx.Done():
		shutdownCtx,cancel:=context.WithTimeout(context.Background(),cfg.Shutdown);defer cancel()
		if err:=server.Shutdown(shutdownCtx);err!=nil{logger.Error("graceful shutdown failed","error",err);_ = server.Close();os.Exit(1)}
		if err:=<-serveErr;err!=nil{logger.Error("HTTP server stopped with error","error",err);os.Exit(1)};logger.Info("HTTP server stopped")
	case err:=<-serveErr:if err!=nil{logger.Error("HTTP server failed","error",err);os.Exit(1)};logger.Error("HTTP server stopped before shutdown was requested");os.Exit(1)
	}
}
func NewServer(cfg config.Config,version string,readiness httpapi.ReadinessCheck,roomStyles ...http.Handler)*http.Server{
	return &http.Server{Addr:cfg.HTTPAddr,Handler:httpapi.NewHandler(version,readiness,roomStyles...),ReadHeaderTimeout:cfg.ReadHeader,ReadTimeout:cfg.Read,WriteTimeout:cfg.Write,IdleTimeout:cfg.Idle,MaxHeaderBytes:1<<20}
}
