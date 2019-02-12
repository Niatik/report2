package main

import (
	"database/sql"
	//"fmt"
	"os"
	"strconv"
	"github.com/joho/godotenv"
	_ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/tealeg/xlsx"

)

func main() {

}

//TableDescription to output a table structure
type TableDescription struct {
	columnName string
	dataType  string
	maxChars int
}

// Creator interface of device creator 
type Creator interface {
	СreateDevice(model string) Device // Параметризированный Фабричный Метод
	registerDevice(Device)            // Регистрация созданого подукта
}

// Device interface of device
type Device interface {
	SetSerial(serial string)
	SetSheet(sheet xlsx.Sheet)
	Send() error // каждый продукт должно быть можно использовать
}

// TSRV030xl for reading data from Excel sheet (TSRV 030,031,032)
type TSRV030xl struct {
	date string
	vnr  float64
	t1   float64
	t2   float64
	t3   float64
	m1   float64
	m2   float64
	m3   float64
	w    float64
}

// TSRV030db for reading data from database (TSRV 030,031,032)
type TSRV030db struct {
	W6   sql.NullFloat64
	m1   sql.NullFloat64
	m2   sql.NullFloat64
	m3   sql.NullFloat64
	t1   sql.NullFloat64
	t2   sql.NullFloat64
	t3   sql.NullFloat64
	tnar sql.NullFloat64
	tpr  sql.NullFloat64
}

// TSRV030Vzl for storing device data
type TSRV030Vzl struct {
	serial string
	sheet  xlsx.Sheet
}

// SetSerial sets the serial number of the device (TSRV 030,031,032) 
func (self *TSRV030Vzl) SetSerial(serial string) {
	self.serial = serial
}

// SetSheet sets the Excel sheet with data of the device (TSRV 030,031,032)
func (self *TSRV030Vzl) SetSheet(sheet xlsx.Sheet) {
	self.sheet = sheet
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
		query := "SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH FROM  [" + dbName + "].INFORMATION_SCHEMA.COLUMNS "
		query += "WHERE TABLE_NAME = '" + tableOriginalName + "' "
		query += "ORDER BY ORDINAL_POSITION"
		queryCreate := "CREATE TABLE [" + dbName + "].[dbo]." + tableOriginalName + "_orig ("
		querySelect := "SELECT "
		queryInsert := "INSERT INTO [" + dbName + "].[dbo]." +tableOriginalName + "_orig ("
		
		tbl, err := db.Query(query)
		checkErr(err)
		r := 0
		for tbl.Next() {
			err = tbl.Scan(&row.columnName, &row.dataType, &row.maxChars)
			if r > 0 {
				queryCreate += ", "
				querySelect += ", "
				queryInsert += ", "
			}
			queryCreate += "[" + row.columnName + "] "
			if row.dataType == "varchar" {				
				queryCreate += row.dataType + " (" + strconv.Itoa(row.maxChars) + ") NULL"	
			} else {
				queryCreate += "[" + row.dataType + "] NULL"
			}
			querySelect += "[" + row.columnName + "] "
			queryInsert += "[" + row.columnName + "]"
			r++
		}
		queryCreate += ") ON [PRIMARY]"
		querySelect += "FROM [" + dbName + "].[dbo]." + tableOriginalName

		_, err = db.Exec(queryCreate)
		checkErr(err)

		readingRows, err := db.Query(querySelect)		
		checkErr(err)


		////////////////////////////////////////////////////
		columns, err := readingRows.Columns()
		checkErr(err)
	
		// Make a slice for the values
		values := make([]sql.RawBytes, len(columns))
	
		// rows.Scan wants '[]interface{}' as an argument, so we must copy the
		// references into such a slice
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		
		for readingRows.Next() {
			queryInsertValues := "("
			// get RawBytes from data
			err = readingRows.Scan(scanArgs...)
			checkErr(err)
			for i, col := range values {
				// you need to be carefull with the datatypes here
				// check out the docs for details on here
				if i > 0 {
					queryInsertValues += ", "	
				}
				queryInsertValues += "'" + string(col) + "'"
				
			}
			//fmt.Println(queryInsert + ") VALUES " + queryInsertValues + ")")
			_, err = db.Exec(queryInsert + ") VALUES " + queryInsertValues + ")")
			checkErr(err)
		}
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

