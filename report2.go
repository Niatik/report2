package main

import (
	"os"
	"github.com/joho/godotenv"
)

func main() {

}

// LoadConfig load parameters for Database 
func LoadConfig() (string, string, string, string, string, error) {
	err := godotenv.Load()
	pathAccess := os.Getenv("MSACCESS_PATH")
	dbServer := os.Getenv("DB_SERVER")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	return pathAccess, dbServer, dbUser, dbPassword, dbName, err
}

// Check error
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
