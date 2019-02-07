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

//TableDescription to output a table structure
type TableDescription struct {
	columnName string
	dataType  string
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

// CopyTable make original copy table in database
func CopyTable(db *sql.DB, dbName string, tableOriginalName string) (string, error) {
	var err error
	if !ExistTable(db, dbName, tableOriginalName + "_orig") {
		var row TableDescription
		query := "SELECT COLUMN_NAME, DATA_TYPE FROM [" + dbName + "].INFORMATION_SCHEMA.COLUMNS "
		query += "WHERE TABLE_NAME = '" + tableOriginalName + "' "
		query += "ORDER BY ORDINAL_POSITION"
		queryCreate := "CREATE TABLE [" + dbName + "].[dbo]." + tableOriginalName + "_orig ("

		tbl, err := db.Query(query)
		checkErr(err)
		for tbl.Next() {
			err = tbl.Scan(&row.columnName, &row.dataType)
			query += "[" + row.columnName + "] "
			query += "[" + row.dataType + "] NULL,"
		}
		query += ") ON [PRIMARY]"

		_, err = db.Exec(queryCreate)
	}

	return tableOriginalName + "_orig", err;
}

// ExistTable checks if there is a table in the database
func ExistTable(db *sql.DB, dbName string, tableName string) (bool) {
	query := "SELECT 1 FROM [" + dbName + "].[dbo]." + tableName
	_, err := db.Query(query)
	result := true;
	if err!= nil {
		result = false
	}
	return result
}

