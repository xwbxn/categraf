package mysql

import (
	"database/sql"
	"log"

	"flashcat.cloud/categraf/pkg/tagx"
	"flashcat.cloud/categraf/types"
	"github.com/toolkits/pkg/container/list"
)

func (m *MySQL) gatherProcesslistByUser(slist *list.SafeList, ins *Instance, db *sql.DB, globalTags map[string]string) {
	if !ins.GatherProcessListProcessByUser {
		return
	}

	rows, err := db.Query(SQL_INFO_SCHEMA_PROCESSLIST_BY_USER)
	if err != nil {
		log.Println("E! failed to get processlist:", err)
		return
	}

	defer rows.Close()

	labels := tagx.Copy(globalTags)

	for rows.Next() {
		var user string
		var connections int64

		err = rows.Scan(&user, &connections)
		if err != nil {
			log.Println("E! failed to scan rows:", err)
			return
		}

		slist.PushFront(types.NewSample("processlist_processes_by_user", connections, labels, map[string]string{"user": user}))
	}
}
