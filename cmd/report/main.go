package main

import (
	"flag"

	"report2/utils"
	"report2"

	"github.com/tealeg/xlsx"
)

func main() {
	fileName := flag.String("file", "Data.xlsx", "File with specified data.")
	sheetName := flag.String("sheet", "Data", "Sheet with specified data.")
	flag.Parse()

	xlFile, err := xlsx.OpenFile(*fileName)
	checkErr(err)

	sheet := xlFile.Sheet[*sheetName]

	model, err := sheet.Cell(0, 0).String()
	checkErr(err)

	serial, err := sheet.Cell(0, 2).String()
	checkErr(err)

	application, err := sheet.Cell(0, 1).String()
	checkErr(err)

	tableName := CalculateTableName(model, model, application)

	//1. Если таблицы нет, значита создать таблицу.
	if !ExistTable(tableName string) {
		CopyTableStructure(tableName string)
	}

	//2. Перенести из оригинальной таблицы данные, которых еще нет в копии таблицы.
	//3. Посчитать и занести данные.
}