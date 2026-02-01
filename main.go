// Author: Iesley Bezerra dos Santos

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Comandos:
/*
mm create=migration name=nomemigration
- cria os arquivos:
- TIMESTAMPUNIX_nomemigration.up.sql
- TIMESTAMPUNIX_nomemigration.down.sql

mm migration=run
- executa os up.sql das migrations nao executadas

mm migration=revert
- executa o down.sql referente a ultima migration executada
*/

type SettingDB struct {
	Sgbd          string `json:"sgbd"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	User          string `json:"user"`
	Dbname        string `json:"dbname"`
	Password      string `json:"password"`
	MigrationsDir string `json:"migrationsDir"`
	SeedersDir    string `json:"seedersDir"`
}

func findPathFileConfig() (configPath string, err error) {
	var root string

	root, err = os.Getwd()
	if err != nil {
		return
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err2 error) error {
		if err2 != nil {
			return err2
		}

		var reg *regexp.Regexp
		reg, err2 = regexp.Compile("mmconfig.json$")
		if err2 != nil {
			return err2
		}

		if reg.MatchString(path) {
			configPath = path
		}

		return nil
	})

	if err != nil {
		return
	}

	if len(configPath) == 0 {
		err = errors.New("mmconfig.json not found")
		return
	}

	return
}

func loadSettingsDB() (sdb *SettingDB, err error) {
	var configPath string
	configPath, err = findPathFileConfig()
	if err != nil {
		return
	}

	var jsonFile *os.File
	jsonFile, err = os.Open(configPath)
	if err != nil {
		return
	}

	var inBytes []byte
	inBytes, err = ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}

	sdb = &SettingDB{}
	err = json.Unmarshal(inBytes, sdb)
	if err != nil {
		return
	}

	// check
	if sdb.Dbname == "" {
		err = fmt.Errorf("error : atribute Dbname empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.Host == "" {
		err = fmt.Errorf("error : atribute Host empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.Password == "" {
		err = fmt.Errorf("error : atribute Password empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.Port == "" {
		err = fmt.Errorf("error : atribute Port empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.Sgbd == "" {
		err = fmt.Errorf("error : atribute Sgbd empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.User == "" {
		err = fmt.Errorf("error : atribute User empty/not declared in mmconfig.json")
		return nil, err
	}
	if sdb.MigrationsDir == "" {
		fmt.Println("WARNING : atribute migrationsDir empty/not declared in mmconfig.json")
		sdb.MigrationsDir = "/migrations"
		fmt.Println("WARNING : migrationsDir set of default '(PathMmConfig)/migrations'")
	}
	if sdb.SeedersDir == "" {
		fmt.Println("WARNING : atribute seedersDir empty/not declared in mmconfig.json")
		sdb.SeedersDir = "/seeders"
		fmt.Println("WARNING : seedersDir set of default '(PathMmConfig)/seeders'")
	}

	// parse values
	arrayPathsFileConfig := strings.Split(configPath, "/")
	pathFolderBased := strings.Join(arrayPathsFileConfig[:len(arrayPathsFileConfig)-1], "/")

	sdb.MigrationsDir = pathFolderBased + sdb.MigrationsDir
	sdb.SeedersDir = pathFolderBased + sdb.SeedersDir

	return
}

func connectDatabase(sdb SettingDB) (db *sql.DB, err error) {
	stringConnection := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", sdb.Host, sdb.Port, sdb.User, sdb.Dbname, sdb.Password)

	db, err = sql.Open(sdb.Sgbd, stringConnection)
	if err != nil {
		return
	}

	err = db.Ping()
	if err != nil {
		return
	}

	return
}

func queryCheckIfTableExist(db *sql.DB, tableName *string) (exist bool, err error) {
	exist = false

	stringQuery := fmt.Sprintf("SELECT count(to_regclass('%s'));", *tableName)

	var count string
	err = db.QueryRow(stringQuery).Scan(&count)
	if err != nil {
		return
	}

	if count == "1" {
		exist = true
	}

	return
}

func queryCreateTableMigrations(db *sql.DB) (err error) {
	queryCreateTable := "CREATE TABLE IF NOT EXISTS public.t_migrations ( id serial primary key, migration_name text not null)"

	_, err = db.Exec(queryCreateTable)
	if err != nil {
		return
	}

	return
}

func queryCreateTableSeeders(db *sql.DB) (err error) {
	queryCreateTable := "CREATE TABLE IF NOT EXISTS public.t_seeders ( id serial primary key, seeder_name text not null)"

	_, err = db.Exec(queryCreateTable)
	if err != nil {
		return
	}

	return
}

func queryReturnMigrationsName(db *sql.DB) (migrationsNames *[]string, err error) {
	var rows *sql.Rows
	rows, err = db.Query("select migration_name from t_migrations;")
	if err != nil {
		return
	}
	defer rows.Close()

	sliceString := make([]string, 0)
	migrationsNames = &sliceString

	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return
		}

		*migrationsNames = append(*migrationsNames, name)
	}

	return
}

func queryReturnSeedersName(db *sql.DB) (seedersNames *[]string, err error) {
	var rows *sql.Rows
	rows, err = db.Query("select seeder_name from t_seeders;")
	if err != nil {
		return
	}
	defer rows.Close()

	sliceString := make([]string, 0)
	seedersNames = &sliceString

	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return
		}

		*seedersNames = append(*seedersNames, name)
	}

	return
}

func queryInsertMigration(db *sql.DB, migrationName string) (err error) {
	var insertedId uint64
	err = db.QueryRow("INSERT INTO t_migrations (migration_name) VALUES ($1) RETURNING id;", migrationName).Scan(&insertedId)
	if err != nil {
		return
	}
	if insertedId == 0 {
		err = errors.New("something went wrong id inserted is equal to zero")
		return
	}

	return
}

func queryInsertSeeder(db *sql.DB, seederName string) (err error) {
	var insertedId uint64
	err = db.QueryRow("INSERT INTO t_seeders (seeder_name) VALUES ($1) RETURNING id;", seederName).Scan(&insertedId)
	if err != nil {
		return
	}
	if insertedId == 0 {
		err = errors.New("something went wrong id inserted is equal to zero")
		return
	}

	return
}

func queryReturnLastMigration(db *sql.DB) (lastMigration *string, err error) {
	var migration string
	err = db.QueryRow("select * from (select migration_name from t_migrations group by migration_name order by migration_name desc) as x limit 1;").Scan(&migration)
	if err != nil {
		return
	}
	lastMigration = &migration
	return
}

func queryDeleteMigration(db *sql.DB, migrationName string) (err error) {
	_, err = db.Exec("delete from t_migrations where migration_name = $1;", migrationName)
	if err != nil {
		return
	}

	return
}

func executeScriptSql(db *sql.DB, pathScript string) (err error) {
	var f *os.File
	f, err = os.Open(pathScript)
	if err != nil {
		return
	}

	var b []byte
	b, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}

	_, err = db.Exec(string(b))
	if err != nil {
		return
	}

	return
}

func nameFilesMigration(migrationsDir *string) (migrationsFileNames *[]string, err error) {
	var files []fs.FileInfo
	files, err = ioutil.ReadDir(*migrationsDir)
	if err != nil {
		return
	}

	sliceString := make([]string, 0)
	migrationsFileNames = &sliceString

	for _, file := range files {
		*migrationsFileNames = append(*migrationsFileNames, file.Name())
	}

	return
}

func nameFilesSeeder(seedersDir *string) (seedersFileNames *[]string, err error) {
	var files []fs.FileInfo
	files, err = ioutil.ReadDir(*seedersDir)
	if err != nil {
		return
	}

	sliceString := make([]string, 0)
	seedersFileNames = &sliceString

	for _, file := range files {
		*seedersFileNames = append(*seedersFileNames, file.Name())
	}

	return
}

func runMigrations(db *sql.DB, migrationsDir *string) (err error) {
	tableName := "public.t_migrations"
	var tableExist bool
	tableExist, err = queryCheckIfTableExist(db, &tableName)
	if err != nil {
		return
	}

	if !tableExist {
		err = queryCreateTableMigrations(db)
		if err != nil {
			return
		}
		fmt.Println("TABLE t_migrations CREATED")
	}

	var migrationNamesDB *[]string
	migrationNamesDB, err = queryReturnMigrationsName(db)
	if err != nil {
		return
	}

	sort.Strings(*migrationNamesDB)

	var fileNames *[]string
	fileNames, err = nameFilesMigration(migrationsDir)
	if err != nil {
		return
	}

	fileNamesFiltered := make([]string, 0)
	for _, v := range *fileNames {
		nameSplit := strings.Split(v, ".")
		if !(len(nameSplit) == 3) || !(nameSplit[2] == "sql") || !(nameSplit[1] == "up") {
			continue
		}

		fileNamesFiltered = append(fileNamesFiltered, nameSplit[0])
	}

	sort.Strings(fileNamesFiltered)

	anyMigrationCreated := false

	for _, fileName := range fileNamesFiltered {
		contains := false
		for _, migrationName := range *migrationNamesDB {
			if fileName == migrationName {
				contains = true
				break
			}
		}

		if !contains {
			fileNameFull := fmt.Sprintf("%s.up.sql", fileName)

			pathFileNameFull := *migrationsDir + "/" + fileNameFull

			err = executeScriptSql(db, pathFileNameFull)
			if err != nil {
				fmt.Printf("Error in run %s\n", fileNameFull)
				return
			}

			err = queryInsertMigration(db, fileName)
			if err != nil {
				return
			}

			anyMigrationCreated = true

			fmt.Printf("MIGRATION %s.up EXECUTED\n", fileName)
		}
	}

	if !anyMigrationCreated {
		fmt.Println("Neither migration created")
	}

	return
}

func revertMigration(db *sql.DB, migrationsDir *string) (err error) {
	tableName := "public.t_migrations"
	var tableExist bool
	tableExist, err = queryCheckIfTableExist(db, &tableName)
	if err != nil {
		return
	}

	if !tableExist {
		err = queryCreateTableMigrations(db)
		if err != nil {
			return
		}
		fmt.Println("TABLE t_migrations CREATED")
	}

	var lastMigration *string
	lastMigration, err = queryReturnLastMigration(db)
	if err != nil {
		return
	}

	fileLastMigration := fmt.Sprintf("%s.down.sql", *lastMigration)

	pathFileLastMigration := *migrationsDir + "/" + fileLastMigration

	err = executeScriptSql(db, pathFileLastMigration)
	if err != nil {
		fmt.Printf("Error in run %s\n", fileLastMigration)
		return
	}

	err = queryDeleteMigration(db, *lastMigration)
	if err != nil {
		return
	}

	fileLastMigrationNoDotSql := fmt.Sprintf("%s.down", *lastMigration)
	fmt.Printf("MIGRATION %s EXECUTED\n", fileLastMigrationNoDotSql)

	return
}

func revertAllMigration(db *sql.DB, migrationsDir *string) (err error) {
	tableName := "public.t_migrations"
	var tableExist bool
	tableExist, err = queryCheckIfTableExist(db, &tableName)
	if err != nil {
		return
	}

	if !tableExist {
		err = queryCreateTableMigrations(db)
		if err != nil {
			return
		}
		fmt.Println("TABLE t_migrations CREATED")
	}

	for {
		var lastMigration *string
		lastMigration, err = queryReturnLastMigration(db)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				return nil
			}
			return
		}

		fileLastMigration := fmt.Sprintf("%s.down.sql", *lastMigration)

		pathFileLastMigration := *migrationsDir + "/" + fileLastMigration

		err = executeScriptSql(db, pathFileLastMigration)
		if err != nil {
			fmt.Printf("Error in run %s\n", fileLastMigration)
			return
		}

		err = queryDeleteMigration(db, *lastMigration)
		if err != nil {
			return
		}

		fileLastMigrationNoDotSql := fmt.Sprintf("%s.down", *lastMigration)
		fmt.Printf("MIGRATION %s EXECUTED\n", fileLastMigrationNoDotSql)
	}
}

func runSeeders(db *sql.DB, seedersDir *string) (err error) {
	tableName := "public.t_seeders"
	var tableExist bool
	tableExist, err = queryCheckIfTableExist(db, &tableName)
	if err != nil {
		return
	}

	if !tableExist {
		err = queryCreateTableSeeders(db)
		if err != nil {
			return
		}
		fmt.Println("TABLE t_seeders CREATED")
	}

	var seederNamesDB *[]string
	seederNamesDB, err = queryReturnSeedersName(db)
	if err != nil {
		return
	}

	sort.Strings(*seederNamesDB)

	var fileNames *[]string
	fileNames, err = nameFilesSeeder(seedersDir)
	if err != nil {
		return
	}

	fileNamesFiltered := make([]string, 0)
	for _, v := range *fileNames {

		nameSplit := strings.Split(v, ".")
		if !(len(nameSplit) == 2) || !(nameSplit[1] == "sql") {
			continue
		}

		fileNamesFiltered = append(fileNamesFiltered, nameSplit[0])
	}

	sort.Strings(fileNamesFiltered)

	anySeederCreated := false

	for _, fileName := range fileNamesFiltered {
		contains := false
		for _, seederName := range *seederNamesDB {
			if fileName == seederName {
				contains = true
				break
			}
		}

		if !contains {
			fileNameFull := fmt.Sprintf("%s.sql", fileName)

			pathFileNameFull := *seedersDir + "/" + fileNameFull

			err = executeScriptSql(db, pathFileNameFull)
			if err != nil {
				fmt.Printf("Error in run %s\n", fileNameFull)
				return
			}

			err = queryInsertSeeder(db, fileName)
			if err != nil {
				return
			}

			anySeederCreated = true

			fmt.Printf("SEEDER %s.up EXECUTED\n", fileName)
		}
	}

	if !anySeederCreated {
		fmt.Println("Neither seeder created")
	}

	return
}

func main() {
	var err error

	// pull database settings
	var sdb *SettingDB
	sdb, err = loadSettingsDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	// connect with database
	var db *sql.DB
	db, err = connectDatabase(*sdb)
	if err != nil {
		fmt.Println(err)
		return
	}

	// declare all flags
	create := flag.String("create", "null", "Specifies what will be created")
	migration := flag.String("migration", "null", "Specifies what will run on migrations")
	seeder := flag.String("seeder", "null", "Specifies what will run on seeders")
	name := flag.String("name", "null", "Inform the name of something according to the context")
	flag.Parse()

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		fmt.Print(err.Error())
	}

	// indicates if the amounts received matched any business rules
	flagsOk := false

	// case : create migration
	if *create == "migration" && *name != "null" {
		flagsOk = true

		// check if migrationDir exist
		if _, err = os.Stat(sdb.MigrationsDir); os.IsNotExist(err) {
			fmt.Printf("Directory of migrationDir:'%s' does not exist\n", sdb.MigrationsDir)
			return
		}

		// create name of files
		timeNow := time.Now().In(loc)
		nameBase := fmt.Sprintf("%d_%s_%s", timeNow.Unix(), timeNow.Format("02_01_2006_150405"), *name)
		nameUp := fmt.Sprintf("%s.up.sql", nameBase)
		nameDown := fmt.Sprintf("%s.down.sql", nameBase)

		// sleep two miliseconds for ensure that the command does not
		// recreate a file that already exists
		time.Sleep(time.Millisecond * 2)

		// create file migration up
		var f1 *os.File
		f1, err = os.Create(sdb.MigrationsDir + "/" + nameUp)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f1.Close()

		// create file migration down
		var f2 *os.File
		f2, err = os.Create(sdb.MigrationsDir + "/" + nameDown)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f2.Close()

		fmt.Printf("migration %s created\n", nameUp)
		fmt.Printf("migration %s created\n", nameDown)
	} else if *create == "seeder" && *name != "null" {
		// case : create seeder
		flagsOk = true

		// check if seedersDir exist
		if _, err = os.Stat(sdb.SeedersDir); os.IsNotExist(err) {
			fmt.Printf("Directory of seedersDir:'%s' does not exist\n", sdb.SeedersDir)
			return
		}

		// create name of files
		timeNow := time.Now().In(loc)
		name := fmt.Sprintf("%d_%s_%s.sql", timeNow.Unix(), timeNow.Format("02_01_2006_150405"), *name)

		// sleep two miliseconds for ensure that the command does not
		// recreate a file that already exists
		time.Sleep(time.Millisecond * 2)

		// create file seeder
		var f *os.File
		f, err = os.Create(sdb.SeedersDir + "/" + name)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		fmt.Printf("seeder %s created\n", name)
	} else if *migration != "null" {
		switch *migration {
		case "run":
			{
				flagsOk = true

				err = runMigrations(db, &sdb.MigrationsDir)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		case "revert":
			{
				flagsOk = true

				err = revertMigration(db, &sdb.MigrationsDir)
				if err != nil {
					if err.Error() == "sql: no rows in result set" {
						fmt.Println("None migration found")
					} else {
						fmt.Println(err)
					}
					return
				}
			}
		case "revertall":
			{
				flagsOk = true

				err = revertAllMigration(db, &sdb.MigrationsDir)
				if err != nil {
					if err.Error() == "sql: no rows in result set" {
						fmt.Println("None migration found")
					} else {
						fmt.Println(err)
					}
					return
				}
			}
		}
	} else if *seeder != "null" {
		switch *seeder {
		case "run":
			{
				flagsOk = true

				err = runSeeders(db, &sdb.SeedersDir)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}

	if !flagsOk {
		fmt.Println("Incorrect Flags")
	}
}
