package tag

import (
	"errors"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	log "github.com/sirupsen/logrus"
)

// errTagStoreRead is the only read error callers see. Callers pass read errors
// on to API responses, and driver errors can name hosts and keyspaces.
var errTagStoreRead = errors.New("tag store read failed")

// queryRows is SimpleDao.Query that returns read errors: the shared helper
// drops them, so a failed read would look like an empty result.
func queryRows(query string, params ...string) ([]map[string]interface{}, error) {
	rows, err := readRows(query, params...)
	if err != nil {
		// Params stay out of the log: they include device ids.
		log.Errorf("tag store read failed: %v; query: %s", err, query)
		return nil, errTagStoreRead
	}
	return rows, nil
}

func readRows(query string, params ...string) ([]map[string]interface{}, error) {
	cc, ok := ds.GetDatabaseClient().(*ds.CassandraClient)
	if !ok {
		return ds.GetSimpleDao().Query(query, params...)
	}
	args := make([]interface{}, len(params))
	for i, p := range params {
		args[i] = p
	}

	cc.ConcurrentQueries <- true
	defer func() { <-cc.ConcurrentQueries }()

	iter := cc.Query(query, args...).Iter()
	var rows []map[string]interface{}
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		rows = append(rows, row)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return rows, nil
}
