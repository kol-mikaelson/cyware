package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Db *pgxpool.Pool
func Dbconnect(){
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DB")
	dbconnstring := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	var err error

	Db,err = pgxpool.New(context.Background(),dbconnstring)
	if err != nil {
		log.Fatal("Unable to connect %v\n",err)
	}
	if err := Db.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping database %v\n",err)
	}
	fmt.Println("Successfully connected to the database!")

}
func Dbclose(){
	if Db != nil {
		Db.Close()
		fmt.Println("Successfully closed database connection!")
	}
}