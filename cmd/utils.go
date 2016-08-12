package cmd

import (
	"database/sql"
	"fmt"
	"github.com/infomodels/database"
)

func getModelAndVersion(dburi string, searchPath string) (model string, modelVersion string, err error) {
	var (
		db *sql.DB
	)

	db, err = database.OpenDatabase(dburi, searchPath)
	if err != nil {
		return
	}
	defer db.Close()

	// From the version_history table, return the model and
	// model_version from the last 'create tables' entry such that there
	// is no 'drop tables' entry after the final 'create tables' entry.
	query := `
with last_create as
(select * from version_history
where operation = 'create tables'
order by datetime desc
limit 1),
last_drop as
(select * from version_history
where operation = 'drop tables'
order by datetime desc
limit 1)
select model, model_version from
(select last_create.model, last_create.model_version, last_create.datetime as last_create_time, last_drop.datetime as last_drop_time from last_create
left join last_drop on 1 = 1) q
where case when last_drop_time is null then true when last_create_time > last_drop_time then true else false end;
  `
	err = db.QueryRow(query).Scan(&model, &modelVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("Can't determine model and version because no active 'create tables' operation in version_history table (search_path %s)", searchPath)
		} else {
			return
		}
	}

	return
}
