package main

import (
	"database/sql"
	//"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	_ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/tealeg/xlsx"
)

func main() {

}

//TableDescription to output a table structure
type TableDescription struct {
	columnName string
	dataType   string
	maxChars   int
}

// Creator interface of device creator
type Creator interface {
	СreateDevice(model string) Device // Parameterized Factory Method
	registerDevice(Device)            // Registration of the created device
}

// Device interface of device
type Device interface {
	SetSerial(serial string)
	SetSheet(sheet xlsx.Sheet)
	Send() error // Every device should be usable
}

// ConcreteCreator struct for concrete device creator
type ConcreteCreator struct {
	devices []*Device // Produced devices
}

// CreateDevice method to create concrete device
func (concreteCreator *ConcreteCreator) CreateDevice(model string, app string) Device {
	var device Device

	if app == "Arc" {
		switch model {
		case "TV7":
			//device = &TV7Arc{}
		default:
			log.Fatalln("Unknown device")
		}
	} else {
		switch model {
		case "TV7":
			//device = &TV7Vzl{}
		case "VKT7":
			//device = &VKT7Vzl{}
		case "TSRV030":
			device = &TSRV030Vzl{}
		case "TSRV034":
			//device = &TSRV034Vzl{}
		default:
			log.Fatalln("Unknown device")
		}
	}

	concreteCreator.RegisterDevice(device)

	return device
}

// RegisterDevice unnecessary function for registering devices in the creator
func (concreteCreator *ConcreteCreator) RegisterDevice(device Device) {
	concreteCreator.devices = append(concreteCreator.devices, &device)
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
func (device *TSRV030Vzl) SetSerial(serial string) {
	device.serial = serial
}

// SetSheet sets the Excel sheet with data of the device (TSRV 030,031,032)
func (device *TSRV030Vzl) SetSheet(sheet xlsx.Sheet) {
	device.sheet = sheet
}

// Send method
func (device *TSRV030Vzl) Send() error {
	_, dbServer, dbUser, dbPassword, dbName, err := LoadConfig()
	db := ConnectMSSQL(dbServer, dbUser, dbPassword)
	defer db.Close()

	_, err = CopyTable(db, dbName, "Суточная_"+device.serial+"_ТСРВ030031032")
	checkErr(err)

	stmtInsertTSRV, err := db.Prepare("insert into [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032]([ДатаВремя], W6, m1, m2, m3, t1, t2, t3, [Tнар], [Tпр]) VALUES (?,?,?,?,?,?,?,?,?,?)")
	checkErr(err)

	stmtUpdateTSRV, err := db.Prepare("update [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032] set W6=?, m1=?, m2=?, m3=?, t1=?, t2=?, t3=?, [Tнар]=?, [Tпр]=? where [ДатаВремя]>=? and [ДатаВремя]<?")
	checkErr(err)

	var xlRow TSRV030xl
	var dbRow TSRV030db
	maxRow := device.sheet.MaxRow
	var tStart time.Time
	var tFinish time.Time
	for rowNumber := 2; rowNumber < maxRow; rowNumber++ {
		xlRow.date, err = device.sheet.Cell(rowNumber, 0).String()
		checkErr(err)
		xlRow.vnr, err = device.sheet.Cell(rowNumber, 1).Float()
		checkErr(err)
		xlRow.t1, err = device.sheet.Cell(rowNumber, 2).Float()
		checkErr(err)
		xlRow.t2, err = device.sheet.Cell(rowNumber, 3).Float()
		checkErr(err)
		xlRow.t3, err = device.sheet.Cell(rowNumber, 4).Float()
		checkErr(err)
		xlRow.m1, err = device.sheet.Cell(rowNumber, 5).Float()
		checkErr(err)
		xlRow.m2, err = device.sheet.Cell(rowNumber, 6).Float()
		checkErr(err)
		xlRow.m3, err = device.sheet.Cell(rowNumber, 7).Float()
		checkErr(err)
		xlRow.w, err = device.sheet.Cell(rowNumber, 8).Float()
		checkErr(err)

		tStart, _ = time.Parse("01-02-06", xlRow.date)
		tFinish = tStart.AddDate(0, 0, 1)

		if rowNumber == 2 {
			query := "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032]"
			query += " WHERE ДатаВремя < '" + tStart.Format("02.01.2006") + "' ORDER BY ДатаВремя DESC"
			readingPrevRows, err := db.Query(query)
			checkErr(err)
			for readingPrevRows.Next() {
				err = readingPrevRows.Scan(&dbRow.W6, &dbRow.m1, &dbRow.m2, &dbRow.m3, &dbRow.t1, &dbRow.t2, &dbRow.t3, &dbRow.tnar, &dbRow.tpr)
				checkErr(err)
			}
		}

		dbRow.W6.Float64 = dbRow.W6.Float64 + Round(xlRow.w*4186.80174034068)
		dbRow.m1.Float64 = dbRow.m1.Float64 + Round(xlRow.m1*1000)
		dbRow.m2.Float64 = dbRow.m2.Float64 + Round(xlRow.m2*1000)
		dbRow.m3.Float64 = dbRow.m3.Float64 + Round(xlRow.m3*1000)
		dbRow.t1.Float64 = Round(xlRow.t1 * 100)
		dbRow.t2.Float64 = Round(xlRow.t2 * 100)
		dbRow.t3.Float64 = Round(xlRow.t3 * 100)
		dbRow.tnar.Float64 = dbRow.tnar.Float64 + Round(xlRow.vnr*3600)
		dbRow.tpr.Float64 = dbRow.tpr.Float64 + Round((24-xlRow.vnr)*3600)

		query := "SELECT COUNT(*) FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032] WHERE ДатаВремя >= '" + tStart.Format("02.01.2006") + "'"
		query += " AND ДатаВремя < '" + tFinish.Format("02.01.2006") + "'"
		readings, err := db.Query(query)
		checkErr(err)

		countReadings := 0
		for readings.Next() {
			err := readings.Scan(&countReadings)
			checkErr(err)
		}

		if countReadings > 0 {
			_, err = stmtUpdateTSRV.Exec(dbRow.W6.Float64, dbRow.m1.Float64, dbRow.m2.Float64, dbRow.m3.Float64, dbRow.t1.Float64, dbRow.t2.Float64, dbRow.t3.Float64, dbRow.tnar.Float64, dbRow.tpr.Float64, tStart, tFinish)
			checkErr(err)
		} else {
			_, err = stmtInsertTSRV.Exec(tStart, dbRow.W6.Float64, dbRow.m1.Float64, dbRow.m2.Float64, dbRow.m3.Float64, dbRow.t1.Float64, dbRow.t2.Float64, dbRow.t3.Float64, dbRow.tnar.Float64, dbRow.tpr.Float64)
			checkErr(err)
		}
	}

	tStart = tFinish
	tFinish = tStart.AddDate(0, 0, 1)

	query := "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
	query += " WHERE ДатаВремя >= '" + tStart.Format("02.01.2006") + "' ORDER BY ДатаВремя ASC"
	readingRows, err := db.Query(query)

	for readingRows.Next() {
		var dbRowPrev TSRV030db
		var dbRowCurr TSRV030db
	
		query := "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
		query += " WHERE ДатаВремя < '" + tStart.Format("02.01.2006") + "' AND ([W6] IS NOT NULL) ORDER BY ДатаВремя DESC"
		readingPrevRows, err := db.Query(query)
		checkErr(err)

		for readingPrevRows.Next() {
			err = readingPrevRows.Scan(&dbRowPrev.W6, &dbRowPrev.m1, &dbRowPrev.m2, &dbRowPrev.m3, &dbRowPrev.t1, &dbRowPrev.t2, &dbRowPrev.t3, &dbRowPrev.tnar, &dbRowPrev.tpr)
			checkErr(err)
		}

		query = "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
		query += " WHERE ДатаВремя >= '" + tStart.Format("02.01.2006") + "' AND ДатаВремя < '" + tFinish.Format("02.01.2006") + "' AND ([W6] IS NOT NULL)"
		readingCurrRows, err := db.Query(query)
		checkErr(err)

		for readingCurrRows.Next() {
			err = readingCurrRows.Scan(&dbRowCurr.W6, &dbRowCurr.m1, &dbRowCurr.m2, &dbRowCurr.m3, &dbRowCurr.t1, &dbRowCurr.t2, &dbRowCurr.t3, &dbRowCurr.tnar, &dbRowCurr.tpr)
			checkErr(err)
		}

		dbRow.W6.Float64 = dbRow.W6.Float64 + (dbRowCurr.W6.Float64-dbRowPrev.W6.Float64)/4186.80174034068 // Round(xlRow.w*4186.80174034068)
		dbRow.m1.Float64 = dbRow.m1.Float64 + (dbRowCurr.m1.Float64-dbRowPrev.m1.Float64)/1000
		dbRow.m2.Float64 = dbRow.m2.Float64 + (dbRowCurr.m2.Float64-dbRowPrev.m2.Float64)/1000
		dbRow.m3.Float64 = dbRow.m3.Float64 + (dbRowCurr.m3.Float64-dbRowPrev.m3.Float64)/1000
		dbRow.t1.Float64 = dbRowCurr.t1.Float64 / 100
		dbRow.t2.Float64 = dbRowCurr.t2.Float64 / 100
		dbRow.t3.Float64 = dbRowCurr.t3.Float64 / 100
		dbRow.tnar.Float64 = dbRow.tnar.Float64 + (dbRowCurr.tnar.Float64-dbRowPrev.tnar.Float64)/3600
		dbRow.tpr.Float64 = dbRow.tpr.Float64 + (dbRowCurr.tpr.Float64-dbRowPrev.tpr.Float64)/3600

		_, err = stmtUpdateTSRV.Exec(dbRow.W6.Float64, dbRow.m1.Float64, dbRow.m2.Float64, dbRow.m3.Float64, dbRow.t1.Float64, dbRow.t2.Float64, dbRow.t3.Float64, dbRow.tnar.Float64, dbRow.tpr.Float64, tStart, tFinish)
		checkErr(err)
	}
	return err
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
	if !ExistTable(db, dbName, tableOriginalName+"_orig") {
		var row TableDescription
		query := "SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH FROM  [" + dbName + "].INFORMATION_SCHEMA.COLUMNS "
		query += "WHERE TABLE_NAME = '" + tableOriginalName + "' "
		query += "ORDER BY ORDINAL_POSITION"
		queryCreate := "CREATE TABLE [" + dbName + "].[dbo]." + tableOriginalName + "_orig ("
		querySelect := "SELECT "
		queryInsert := "INSERT INTO [" + dbName + "].[dbo]." + tableOriginalName + "_orig ("

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

	return tableOriginalName + "_orig", err
}

// ExistTable checks if there is a table in the database
func ExistTable(db *sql.DB, dbName string, tableName string) bool {
	query := "SELECT 1 FROM [" + dbName + "].[dbo]." + tableName
	_, err := db.Query(query)
	result := true
	if err != nil {
		result = false
	}
	return result
}

// Round - number rounding function
func Round(f float64) float64 {
	return math.Floor(f + .5)
}
