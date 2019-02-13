package tsrv030

import (
	"database/sql"
	//"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/alexbrainman/odbc"
	"github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/tealeg/xlsx"
)

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

// TSRV030 for storing device data
type TSRV030 struct {
	serial string
	sheet  xlsx.Sheet
}

// SetSerial sets the serial number of the device (TSRV 030,031,032)
func (device *TSRV030) SetSerial(serial string) {
	device.serial = serial
}

// SetSheet sets the Excel sheet with data of the device (TSRV 030,031,032)
func (device *TSRV030) SetSheet(sheet xlsx.Sheet) {
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
			query += " WHERE ДатаВремя < '" + TimeToString(tStart) + "' ORDER BY ДатаВремя DESC"
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

		query := "SELECT COUNT(*) FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032] WHERE ДатаВремя >= '" + TimeToString(tStart) + "'"
		query += " AND ДатаВремя < '" + TimeToString(tFinish) + "'"
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
	
	query := "SELECT [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
	query += " WHERE ДатаВремя >= '" + TimeToString(tStart) + "' ORDER BY ДатаВремя ASC"
	readingRows, err := db.Query(query)

	for readingRows.Next() {
		var dbRowPrev TSRV030db
		var dbRowCurr TSRV030db
	
		query := "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
		query += " WHERE ДатаВремя < '" + TimeToString(tStart) + "' AND ([W6] > 0) ORDER BY ДатаВремя DESC"
		readingPrevRows, err := db.Query(query)
		checkErr(err)

		for readingPrevRows.Next() {
			err = readingPrevRows.Scan(&dbRowPrev.W6, &dbRowPrev.m1, &dbRowPrev.m2, &dbRowPrev.m3, &dbRowPrev.t1, &dbRowPrev.t2, &dbRowPrev.t3, &dbRowPrev.tnar, &dbRowPrev.tpr)
			checkErr(err)
		}

		query = "SELECT TOP 1 [W6], [m1], [m2], [m3], [t1], [t2], [t3], [Tнар], [Tпр]	FROM [" + dbName + "].[dbo].[Суточная_" + device.serial + "_ТСРВ030031032_orig]"
		query += " WHERE ДатаВремя >= '" + TimeToString(tStart) + "' AND ДатаВремя < '" + TimeToString(tFinish) + "' AND ([W6] > 0)"
		readingCurrRows, err := db.Query(query)
		checkErr(err)

		for readingCurrRows.Next() {
			err = readingCurrRows.Scan(&dbRowCurr.W6, &dbRowCurr.m1, &dbRowCurr.m2, &dbRowCurr.m3, &dbRowCurr.t1, &dbRowCurr.t2, &dbRowCurr.t3, &dbRowCurr.tnar, &dbRowCurr.tpr)
			checkErr(err)
		}

		if dbRowCurr.W6.Float64 > 0 {
			dbRow.W6.Float64 = dbRow.W6.Float64 + (dbRowCurr.W6.Float64-dbRowPrev.W6.Float64)
			dbRow.m1.Float64 = dbRow.m1.Float64 + (dbRowCurr.m1.Float64-dbRowPrev.m1.Float64)
			dbRow.m2.Float64 = dbRow.m2.Float64 + (dbRowCurr.m2.Float64-dbRowPrev.m2.Float64)
			dbRow.m3.Float64 = dbRow.m3.Float64 + (dbRowCurr.m3.Float64-dbRowPrev.m3.Float64)
			dbRow.t1.Float64 = dbRowCurr.t1.Float64
			dbRow.t2.Float64 = dbRowCurr.t2.Float64
			dbRow.t3.Float64 = dbRowCurr.t3.Float64
			dbRow.tnar.Float64 = dbRow.tnar.Float64 + (dbRowCurr.tnar.Float64-dbRowPrev.tnar.Float64)
			dbRow.tpr.Float64 = dbRow.tpr.Float64 + (dbRowCurr.tpr.Float64-dbRowPrev.tpr.Float64)
		}
		if dbRowCurr.W6.Float64 > 0 {
			_, err = stmtUpdateTSRV.Exec(dbRow.W6.Float64, dbRow.m1.Float64, dbRow.m2.Float64, dbRow.m3.Float64, dbRow.t1.Float64, dbRow.t2.Float64, dbRow.t3.Float64, dbRow.tnar.Float64, dbRow.tpr.Float64, tStart, tFinish)
		} else {
			_, err = stmtUpdateTSRV.Exec(nil, nil, nil, nil, nil, nil, nil, nil, nil, tStart, tFinish)
		}
		checkErr(err)

		tStart = tFinish
		tFinish = tStart.AddDate(0, 0, 1)	
	}
	return err
}

