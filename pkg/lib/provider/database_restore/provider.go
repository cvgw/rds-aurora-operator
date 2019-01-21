package database_restore

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func RestoreSQLFile(sqlFile string) {
	pathSplit := strings.Split(sqlFile, "/")
	sqlDirectory := strings.Join(pathSplit[:len(pathSplit)-1], "/")

	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("could not read sql file: %s", err)
	}
	splitSql := strings.Split(string(sqlContent), ";")

	db, err := sql.Open("mysql", "admin:123456@/")
	var f func([]string)
	f = func(sqlStatements []string) {
		for i, statement := range sqlStatements {
			re := regexp.MustCompile(`\r?\n`)
			statement = re.ReplaceAllString(statement, "")

			re = regexp.MustCompile(`^source.*`)
			match := re.MatchString(statement)
			if match {
				file := strings.Replace(statement, "source", "", 1)
				file = strings.TrimSpace(file)

				sourceContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", sqlDirectory, file))
				if err != nil {
					log.Fatalf("could not read source file %s %s", file, err)
				}

				f(strings.Split(string(sourceContent), ";"))
				continue
			}

			if statement == "" {
				continue
			}

			if i%10 == 0 {
				log.Info("sql executing")
			}

			_, err := db.Exec(statement)
			if err != nil {
				log.Fatalf("could not execute statement %s %s", statement, err)
			}
		}
	}

	f(splitSql)
}
