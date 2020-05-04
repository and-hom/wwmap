package queries

//go:generate go-bindata -pkg $GOPACKAGE -o bindata.go ./

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"regexp"
	"strings"
)

const SEPARATOR_RE = "--@([\\w-]+)"

var separatorRe = regexp.MustCompile(SEPARATOR_RE)
var cache = make(map[string]map[string]string)

func read(r []byte, result *map[string]string) {
	data := strings.Split(string(r), "\n")
	var key string
	var contents *bytes.Buffer

	for _, line := range data {
		found := separatorRe.FindStringSubmatch(line)
		if len(found) >= 2 && found[1] != "" {
			if key != "" {
				(*result)[key] = contents.String()
			}
			key = found[1]
			contents = bytes.NewBufferString("")
		} else {
			if key == "" {
				log.Fatalf("Should use --@query-name construction before query: %s", line)
			}
			if contents.Len() > 0 {
				contents.WriteString(" ")
			}
			contents.WriteString(line)
		}
	}
	if key != "" {
		(*result)[key] = contents.String()
	}
}

func getQueriesOfFile(file string) map[string]string {
	queriesOfFile, foundInCache := cache[file]
	if !foundInCache {
		log.Debug("Not in cache: ", file)
		queriesOfFile = make(map[string]string)
		cache[file] = queriesOfFile

		sqlFileBytes, err := Asset(file + ".sql")
		if err != nil {
			log.Fatalf("Can not load sql queries file %s: %s", file, err.Error())
		}

		read(sqlFileBytes, &queriesOfFile)
	} else {
		log.Debug("In cache: ", file)
	}
	return queriesOfFile
}

const SUB_QUERY_REPLACE = "___(.*?)___"

var subQueryReplacer = regexp.MustCompile(SUB_QUERY_REPLACE)

func sqlQuery(file string, name string, walkedIdsStack []string) string {
	queriesOfFile := getQueriesOfFile(file)
	query, found := queriesOfFile[name]
	if !found {
		log.Fatalf("Can not get sql query for key %s in file %s", name, file)
	}
	query = strings.Replace(query, "\n", "", -1)
	log.Debug("\"" + query + "\"")
	return subQueryReplacer.ReplaceAllStringFunc(query, func(src string) string {
		queryId := subQueryReplacer.FindStringSubmatch(src)[1]
		for i := 0; i < len(walkedIdsStack); i++ {
			if walkedIdsStack[i] == queryId {
				log.Fatalf("Can not replace placeholder %s to query: cyclic dependency detected. Replacement stack is: %v", src, walkedIdsStack)
			}
		}
		return sqlQuery(file, queryId, append(walkedIdsStack, queryId))
	})
}

func SqlQuery(file string, name string) string {
	return sqlQuery(file, name, []string{name})
}
