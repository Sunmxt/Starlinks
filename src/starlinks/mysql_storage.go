package starlinks

import (
	"database/sql"
	"errors"
)

const (
	SQL_LINK_MAP = `
        CREATE TABLE IF NOT EXIST LINKS(
            ID BIGINT PRIMARY KEY UNIQUE NOT NULL,
            URL VARCHAR(32) UNIQUE NOT NULL, 
            COUNT BIGINT NOT NULL,
        )`

	SQL_REQ_LOG = `
        CREATE TABLE IF NOT EXIST REQ_LOG(
            ID BIGINT PRIMARY KEY UNIQUE NOT NULL,
            REF_ID BIGINT UNIQUE NOT NULL,
            TIME VARCHAR(16) NOT NULL,
            DEVICE VARCHAR(32) NOT NULL,
            IP VARCHAR(32) NOT NULL,
            EXTRA VARCHAR(16) NOT NULL,
            FOREIGN KEY(REF_ID) REFERENCES LINKS(ID)
        )
    `

	SQL_QUERY_LINK  = "SELECT URL FROM LINKS WHERE (ID = %s)"
	SQL_UPDATE_CNT  = "UPDATE LINKS SET COUNT=COUNT+1 WHERE (ID = %s)"
	SQL_ADD_LINK    = "INSERT INTO LINKS(URL) VALUE (%s)"
	SQL_REMOVE_LINK = "DELETE FROM LINKS WHERE (ID = %s)"
)

type MySQLLinkStorage struct {
	db *sql.DB
}

func NewMySQLLinkStorage(dsn string) (LinkStorage, error) {
	sql, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ls := &MySQLLinkStorage{
		db: sql,
	}
	ls.InitTables()

	return ls, nil
}

func (sto *MySQLLinkStorage) QueryLink(id LinkID) (string, error) {
	var row *sql.Row
	var err error
	var url string

	row = sto.db.QueryRow(SQL_LINK_MAP, id.ToString())

	if err != nil || row.Scan(&url) != nil {
		return "", err
	}
	if url == "" {
		return "", nil
	}

	if _, err = sto.db.Exec(SQL_UPDATE_CNT, id.ToString()); err != nil {
		return "", err
	}

	return url, nil
}

func (sto *MySQLLinkStorage) QueryLinks(ids []LinkID) ([]string, error) {
	var err error
	var tx *sql.Tx
	var row *sql.Row

	if len(ids) == 0 {
		return make([]string, 0), nil
	}
	if len(ids) < 0 {
		return nil, errors.New("Invalid ID List.")
	}

	if tx, err = sto.db.Begin(); err != nil {
		return nil, err
	}

	urls := make([]string, len(ids))
	for i, id := range ids {
		row = sto.db.QueryRow(SQL_QUERY_LINK, id)
		err = row.Scan(&urls[i])
		if err != nil {
			urls[i] = ""
		}
	}
	tx.Commit()

	// Update count
	if tx, err = sto.db.Begin(); err != nil {
		return urls, err
	}
	for i, url := range urls {
		if url == "" {
			continue
		}
		tx.Exec(SQL_UPDATE_CNT, ids[i])
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
	}

	return urls, err
}

func (sto *MySQLLinkStorage) AddLink(url string) (LinkID, error) {
	var err error
	var result sql.Result
	var new_id int64

	result, err = sto.db.Exec(SQL_ADD_LINK, url)
	if err != nil {
		return 0, err
	}
	new_id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return LinkID(new_id), nil
}

func (sto *MySQLLinkStorage) AddLinks(urls []string) ([]LinkID, error) {
	var err error
	var result sql.Result
	var tx *sql.Tx
	var new_id int64

	if tx, err = sto.db.Begin(); err != nil {
		return nil, err
	}
	ids := make([]LinkID, len(urls))
	for i, url := range urls {
		result, err = tx.Exec(SQL_ADD_LINK, url)
		if err != nil {
			ids[i] = 0
		} else {
			if new_id, err = result.LastInsertId(); err != nil {
				ids[i] = LinkID(new_id)
			} else {
				ids[i] = 0
			}
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return ids, nil
}

func (sto *MySQLLinkStorage) RemoveLink(id LinkID) error {
	var err error
	var result sql.Result
	var affected int64

	if result, err = sto.db.Exec(SQL_REMOVE_LINK, id); err != nil {
		return err
	}
	if affected, err = result.RowsAffected(); err != nil {
		return err
	}
	if affected < 1 {
		return errors.New("Invalid link id.")
	}

	return nil
}

func (sto *MySQLLinkStorage) RemoveLinks(ids []LinkID) error {
	var err error
	var tx *sql.Tx

	if tx, err = sto.db.Begin(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	for _, id := range ids {
		if _, err = tx.Exec(SQL_REMOVE_LINK, id); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
	}

	return err
}

func (sto *MySQLLinkStorage) InitTables() error {
	var err error
	var tx *sql.Tx

	tx, err = sto.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Exec(SQL_LINK_MAP); err != nil {
		return err
	}
	if _, err = tx.Exec(SQL_REQ_LOG); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
