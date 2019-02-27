package utils_test

import (
    "testing"

    . "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "report2/utils"
)

func TestUtils(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Report Utilities Suite")
}

var _ = Describe("Report Utils", func() {
	
	Context("Copy Table", func() {
		_, dbServer, dbUser, dbPassword, dbName, _ := LoadConfig()
		db := ConnectMSSQL(dbServer, dbUser, dbPassword)
		tableName := "Узлы"	
		It("can copy table in database", func() {
			tableCopyName, err := CopyTable(db, dbName, tableName)
			Expect(tableCopyName).Should(Equal(tableName + "_orig"))
			Expect(err).Should(BeNil())

		})
	})

	Context("Exist Table", func() {
		_, dbServer, dbUser, dbPassword, dbName, _ := LoadConfig()
		db := ConnectMSSQL(dbServer, dbUser, dbPassword)
		tableName := "Узлы"
		It("has the table in database", func() {
			Expect(ExistTable(db, dbName, tableName)).Should(BeTrue())
		})	
	})

	Context("Connect MS SQL Database", func() {
		_, dbServer, dbUser, dbPassword, _, _ := LoadConfig()
		db := ConnectMSSQL(dbServer, dbUser, dbPassword)
		It("can connect to database", func() {
			Expect(db).ShouldNot(BeNil())
		})
	})

	Context("Connect MS Access Database", func() {
		pathAccess, _, _, _, _, _ := LoadConfig()
		db := ConnectMSACCESS(pathAccess)
		It("can connect to database", func() {
			Expect(db).ShouldNot(BeNil())
		})
	})

	Context("Load Configuration", func() {
		dbServer := ""
		dbUser := ""
		dbPassword := ""
		dbName := ""
		pathAccess, dbServer, dbUser, dbPassword, dbName, err := LoadConfig()
		It("can load configuration from '.env' file", func() {
			Expect(err).Should(BeNil())
			Expect(dbServer).ShouldNot(BeEmpty())
			Expect(dbUser).ShouldNot(BeEmpty())
			Expect(dbPassword).ShouldNot(BeEmpty())
			Expect(dbName).ShouldNot(BeEmpty())
			Expect(pathAccess).ShouldNot(BeEmpty())
		})
	})
})