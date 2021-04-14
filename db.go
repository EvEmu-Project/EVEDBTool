package main

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

func getDB() *sql.DB { //Create database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true&maxAllowedPacket=0", viper.GetString("db-user"), viper.GetString("db-pass"), viper.GetString("db-host"), viper.GetString("db-port"), viper.GetString("db-database"))
	dialect := "mysql"
	db, err := sql.Open(dialect, dsn)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func checkDBConnection() {
	tries := 0
	for tries < 1000 {
		db := getDB()
		_, err := db.Query("SELECT 1") //Query doesn't need to do anything, we just use to see if DB is not dead
		if err == nil {
			db.Close()
			log.Info("DB connection re-established.")
			break
		}
		db.Close()
		log.Info("DB connection died, writing for server...")
		time.Sleep(2 * time.Second)
		tries += 1
	}
}

func GetNumberOfTables() int {
	db := getDB()
	var value string
	sqlStatement := `SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?;`
	err := db.QueryRow(sqlStatement, viper.GetString("db-database")).Scan(&value)
	if err != nil {
		log.Fatal("Failed to query db; ", err)
	}

	number, _ := strconv.Atoi(value)

	db.Close()
	return number
}

//Read in gzipped file into an in-memory string array
func ReadGzFile(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer gr.Close()

	var lines []string
	scanner := bufio.NewScanner(gr)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "--") { //We don't want comments
			if strings.TrimSpace(scanner.Text()) != "" { //Or blank lines
				if scanner.Text()[len(scanner.Text())-1:] == ";" {
					newString := scanner.Text()[:len(scanner.Text())-1] + "[LINEBREAK]"
					lines = append(lines, newString)
				} else {
					lines = append(lines, scanner.Text())
				}
			}
		}
	}
	output := strings.Join(lines, "")

	var queryset []string
	for _, str := range strings.Split(output, "[LINEBREAK]") {
		if str != "" {
			queryset = append(queryset, str)
		}
	}

	for i, s := range queryset {
		log.Trace(i, " ##", s, "##")
	}

	return queryset, scanner.Err()
}

func BuildMigration(fileName string, data []string) *migrate.Migration {
	//Build our migration for base file
	log.Debug("Building migration for base file ", fileName)

	migration := &migrate.Migration{
		Id:   "BASE_" + fileName,
		Up:   data,
		Down: []string{},
	}
	return migration
}

func InstallBase() {

	var files []string

	//Walk base dir for all files and put that into an array
	err := filepath.Walk(viper.GetString("base-dir"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() { //We only want files, not dirs
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Error("ERROR: ", err)
	}
	log.Trace("Base files to install: ")
	for _, file := range files {
		log.Trace(file)
	}
	migrations := []*migrate.Migration{}
	migrate.SetTable("base_migrations")

	for _, file := range files {
		log.Debug("Decompressing ", file, "...")

		//Un-gzip the files...
		fileData, err := ReadGzFile(file)
		if err != nil {
			log.Error("Failed to read in GZ data: ", err)
		}

		log.Info("Building migration for ", file, "...")
		newMigration := BuildMigration(file, fileData)
		migrations = append(migrations, newMigration)

		log.Info("Executing migration...")
		migrationSource := &migrate.MemoryMigrationSource{migrations}

		//Create a new DB connection (to avoid exhausting limit)
		db := getDB()
		n, err := migrate.Exec(db, "mysql", migrationSource, migrate.Up)
		if err != nil {
			log.Error("Error installing migration: ", err)
			//Check if DB died
			checkDBConnection()
		}
		db.Close()
		log.Info("Applied ", n, " migrations!")
	}
}

func InstallMigrations() {
	// OR: Read migrations from a folder:
	migrationSource := &migrate.FileMigrationSource{
		Dir: viper.GetString("migrations-dir"),
	}
	migrate.SetTable("migrations")
	//Create a new DB connection (to avoid exhausting limit)
	db := getDB()

	n, err := migrate.Exec(db, "mysql", migrationSource, migrate.Up)
	if err != nil {
		log.Error("Error installing migration: ", err)
	}
	db.Close()
	log.Info("Applied %d migrations!\n", n)
}
