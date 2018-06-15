package queries

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
			contents.WriteString(line)
			contents.WriteString(" ")
		}
	}
	if key != "" {
		(*result)[key] = contents.String()
	}
}

func SqlQuery(file string, name string) string {
	b, err := Asset(file + ".sql")
	if err != nil {
		log.Fatalf("Can not find query %s in file %s: %s", name, file, err.Error())
	}

	queriesOfFile, foundInCache := cache[file]
	if !foundInCache {
		log.Debug("Not in cache: ", file)
		queriesOfFile = make(map[string]string)
		cache[file] = queriesOfFile
	} else {
		log.Debug("In cache: ", file)
	}

	read(b, &queriesOfFile)
	query, found := queriesOfFile[name]
	if !found {
		log.Fatalf("Can not get sql query for key %s in file %s", name, file)
	}
	return query
}
