package main

import (
	"database/sql"
	"os"
	"github.com/joho/godotenv"
	_ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"

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

// ConnectMSSQL connect to MS SQL Server
func ConnectMSSQL(dbServer, dbUser, dbPassword string) *sql.DB {
	db, err := sql.Open("mssql", "server="+dbServer+";user id="+dbUser+";password="+dbPassword)
	checkErr(err)
	return db
}

// ConnectMSACCESS connect to MS Access Database
func ConnectMSACCESS(pathAccess string) *sql.DB {
	connAccess := "Driver=Microsoft Access Driver (*.mdb, *.accdb);driverid=25;DBQ=" + pathAccess + ";FIL=MS Access;SafeTransaction=0"
	dbAccess, err := sql.Open("odbc", connAccess)
	checkErr(err)
	return dbAccess
}

