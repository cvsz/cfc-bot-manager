package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main(){
	dsn:=os.Getenv("DATABASE_URL");if dsn==""{log.Fatal("DATABASE_URL is required")}
	db,err:=sql.Open("pgx",dsn);if err!=nil{log.Fatalf("open database: %v",err)};defer db.Close()
	ctx,cancel:=context.WithTimeout(context.Background(),30*time.Second);defer cancel()
	if err:=db.PingContext(ctx);err!=nil{log.Fatalf("database unavailable: %v",err)}
	if err:=messagestyle.ApplyMigrations(ctx,db);err!=nil{log.Fatalf("apply migrations: %v",err)}
	log.Print("database migrations applied")
}
