package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MariaDB struct {
	Username       string
	Unixsocketpath string
	Password       string
	Host           string
	Port           string
	Database       string
	Charset        string
}

func (db *MariaDB) DriverName() string {
	return "mysql"
}

// ConnectionString returns the connection string. Possible formats:
//  - user@unix(/path/to/socket)/dbname?charset=utf8
//  - user:password@/dbname
//  - user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname
//  - user:password@tcp(localhost:5555)/dbname?charset=utf8
func (db *MariaDB) ConnectionString() string {
	var connString string

	if "" == db.Username || "" == db.Database {
		return ""
	}
	if "" == db.Password && "" == db.Unixsocketpath {
		return ""
	}
	if "" != db.Unixsocketpath {
		connString = db.Username + "@unix(" + db.Unixsocketpath + ")/" + db.Database
		if "" != db.Charset {
			connString += "?charset=" + db.Charset
		}
		return connString
	}
	if "" == db.Host || "" == db.Port {
		connString = db.Username + ":" + db.Password + "@/" + db.Database
		if "" != db.Charset {
			connString += "?charset=" + db.Charset
		}
		return connString
	}

	connString = db.Username + ":" + db.Password + "@tcp(" + db.Host + ":" + db.Port + ")/" + db.Database
	if "" != db.Charset {
		connString += "?charset=" + db.Charset
	}

	return connString
}

func (db *MariaDB) Test() error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())
	if nil != err {
		return err
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		return err
	}

	version := ""
	conn.QueryRow("SELECT VERSION()").Scan(&version)
	if "" == version {
		return errors.New("Couldn't get SQL version.")
	}

	return nil
}

func (db *MariaDB) WriteApplicant(battleid int, battletag, character, email, realName, location, age, gender, computerSpecs,
	previousGuilds, reasonsLeavingGuilds, whyJoinThisGuild, references, finalRemarks string) error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	statement := `INSERT INTO applications(battleid, status, battletag, wowcharacter, email, realname, location,
	age, gender, computerspecs, previousguilds, reasonsleavingguilds, whyjointhisguild, 
	wowreferences, finalremarks) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	in, err := conn.Prepare(statement)
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec(battleid, 2, battletag, character, email, realName, location, age, gender, computerSpecs,
		previousGuilds, reasonsLeavingGuilds, whyJoinThisGuild, references, finalRemarks)

	return nil
}

func (db *MariaDB) GetApplicant(id int) (bool, error) {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return false, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM applications WHERE id = ?", id)

	if nil != err {
		return false, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		count++
	}

	err = rows.Err()

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AcceptApplicant accepts an applicant by setting their status to 1(Accepted).
func (db *MariaDB) AcceptApplicant(id int) error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	in, err := conn.Prepare("UPDATE applications SET status = 1 WHERE id = ?")
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec(id)

	return nil
}

// RejectApplicant rejects an applicant by setting their status to 0(Rejected).
func (db *MariaDB) RejectApplicant(id int) error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	in, err := conn.Prepare("UPDATE applications SET status = 0 WHERE id = ?")
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec(id)

	return nil
}

func (db *MariaDB) WriteNewsPost(title, body, author string) error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	in, err := conn.Prepare("INSERT INTO newsposts(title, body, date, author) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec(title, body, time.Now(), author)

	return nil
}

func (db *MariaDB) ReadNewsPosts() ([]int, []string, []string, []time.Time, []string, error) {
	var (
		id     int
		title  string
		body   string
		date   time.Time
		author string

		ids     []int
		titles  []string
		bodys   []string
		dates   []time.Time
		authors []string
	)

	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return nil, nil, nil, nil, nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT * from newsposts")

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &title, &body, &date, &author)

		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		ids = append(ids, id)
		titles = append(titles, title)
		bodys = append(bodys, body)
		dates = append(dates, date)
		authors = append(authors, author)
	}

	err = rows.Err()

	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return ids, titles, bodys, dates, authors, nil
}

func (db *MariaDB) Install() error {
	fmt.Printf("Creating table '%s.Applications...\n", db.Database)
	err := db.createTableApplications()

	if nil != err {
		return err
	}
	fmt.Printf("Successfully created table '%s.Applications\n", db.Database)

	fmt.Printf("Creating table '%s.NewsPosts...\n", db.Database)
	err = db.createTableNewsPosts()

	if nil != err {
		return err
	}
	fmt.Printf("Successfully created table '%s.NewsPosts\n", db.Database)

	return nil
}

func (db *MariaDB) createTableApplications() error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	statement := `CREATE TABLE applications(
		id BIGINT NOT NULL AUTO_INCREMENT,
		status INT NOT NULL,
		battleid BIGINT NOT NULL,
		battletag varchar(50) NOT NULL,
		wowcharacter varchar(50) NOT NULL,
		email varchar(50) NOT NULL, 
		realname varchar(50) NOT NULL, 
		location varchar(100) NOT NULL, 
		age TINYINT NOT NULL, 
		gender varchar(20) NOT NULL, 
		computerspecs varchar(500) NOT NULL, 
		previousguilds varchar(500) NOT NULL, 
		reasonsleavingguilds varchar(500) NOT NULL, 
		whyjointhisguild varchar(500) NOT NULL, 
		wowreferences varchar(500) NOT NULL, 
		finalremarks varchar(500) NOT NULL, 
		PRIMARY KEY (id)
		) ENGINE = InnoDB;`

	in, err := conn.Prepare(statement)
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec()

	return nil
}

func (db *MariaDB) createTableNewsPosts() error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	statement := `CREATE TABLE newsposts(
		id BIGINT NOT NULL AUTO_INCREMENT, 
		title VARCHAR(128) NOT NULL, 
		body MEDIUMBLOB NOT NULL, 
		date DATETIME NOT NULL, 
		author VARCHAR(64) NOT NULL, 
		PRIMARY KEY (id)
		) ENGINE = InnoDB;`

	in, err := conn.Prepare(statement)
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec()

	return nil
}

func (db *MariaDB) createTableNewsPostComments() error {
	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if nil != err {
		return err
	}
	defer conn.Close()

	statement := `CREATE TABLE newscomments(
		id BIGINT NOT NULL AUTO_INCREMENT, 
		pid BIGINT NOT NULL, 
		title VARCHAR(128) NOT NULL, 
		body MEDIUMBLOB NOT NULL, 
		date DATETIME NOT NULL, 
		author VARCHAR(64) NOT NULL, 
		PRIMARY KEY (id), 
		CONSTRAINT ` + "`fk_parent_post`" + ` 
		FOREIGN KEY (pid) REFERENCES newsposts (id) 
		ON DELETE CASCADE ON UPDATE RESTRICT
		) ENGINE = InnoDB;`

	in, err := conn.Prepare(statement)
	if err != nil {
		return err
	}
	defer in.Close()

	in.Exec()

	return nil
}

func (db *MariaDB) ViewApplicant(id int) ([]int, []string, []string, []string, []string, []string, []string, []string, []string,
	[]string, []string, []string, []string, []string, error) {
	var (
		a int
		b string
		c string
		d string
		e string
		f string
		g string
		h string
		i string
		j string
		k string
		l string
		m string
		n string

		as []int
		bs []string
		cs []string
		ds []string
		es []string
		fs []string
		gs []string
		hs []string
		is []string
		js []string
		ks []string
		ls []string
		ms []string
		ns []string
	)

	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM applications WHERE id = ?", id)

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&a, &b, &c, &d, &e, &f, &g, &h, &i, &j, &k, &l, &m, &n)

		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}

		as = append(as, a)
		bs = append(bs, b)
		cs = append(cs, c)
		ds = append(ds, d)
		es = append(es, e)
		fs = append(fs, f)
		gs = append(gs, g)
		hs = append(hs, h)
		is = append(is, i)
		js = append(js, j)
		ks = append(ks, k)
		ls = append(ls, l)
		ms = append(ms, m)
		ns = append(ns, n)
	}

	err = rows.Err()

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	return as, bs, cs, ds, es, fs, gs, hs, is, js, ks, ls, ms, ns, nil
}

func (db *MariaDB) ViewAllApplicants() ([]int, []string, []string, []string, []string, []string, []string, []string, []string,
	[]string, []string, []string, []string, []string, error) {
	var (
		a int
		b string
		c string
		d string
		e string
		f string
		g string
		h string
		i string
		j string
		k string
		l string
		m string
		n string

		as []int
		bs []string
		cs []string
		ds []string
		es []string
		fs []string
		gs []string
		hs []string
		is []string
		js []string
		ks []string
		ls []string
		ms []string
		ns []string
	)

	conn, err := sql.Open(db.DriverName(), db.ConnectionString())

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM applications")

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&a, &b, &c, &d, &e, &f, &g, &h, &i, &j, &k, &l, &m, &n)

		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}

		as = append(as, a)
		bs = append(bs, b)
		cs = append(cs, c)
		ds = append(ds, d)
		es = append(es, e)
		fs = append(fs, f)
		gs = append(gs, g)
		hs = append(hs, h)
		is = append(is, i)
		js = append(js, j)
		ks = append(ks, k)
		ls = append(ls, l)
		ms = append(ms, m)
		ns = append(ns, n)
	}

	err = rows.Err()

	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	return as, bs, cs, ds, es, fs, gs, hs, is, js, ks, ls, ms, ns, nil
}

/*
SQL Queries

CREATE DATABASE guild;

USE guild;

CREATE TABLE applications(id BIGINT NOT NULL AUTO_INCREMENT, battleid BIGINT NOT NULL, battletag varchar(50) NOT NULL,wowcharacter varchar(50) NOT NULL, email varchar(50) NOT NULL, realname varchar(50) NOT NULL, location varchar(100) NOT NULL, age TINYINT NOT NULL, gender varchar(20) NOT NULL, computerspecs varchar(500) NOT NULL, previousguilds varchar(500) NOT NULL, reasonsleavingguilds varchar(500) NOT NULL, whyjointhisguild varchar(500) NOT NULL, wowreferences varchar(500) NOT NULL, finalremarks varchar(500) NOT NULL, PRIMARY KEY (id)) ENGINE = InnoDB;

CREATE TABLE newsposts(id BIGINT NOT NULL AUTO_INCREMENT, title VARCHAR(128) NOT NULL, body MEDIUMBLOB NOT NULL, date DATETIME NOT NULL, author VARCHAR(64) NOT NULL, PRIMARY KEY (id)) ENGINE = InnoDB;

CREATE TABLE newscomments(id BIGINT NOT NULL AUTO_INCREMENT, pid BIGINT NOT NULL, title VARCHAR(128) NOT NULL, body MEDIUMBLOB NOT NULL, date DATETIME NOT NULL, author VARCHAR(64) NOT NULL, PRIMARY KEY (id), CONSTRAINT `fk_parent_post` FOREIGN KEY (pid) REFERENCES newsposts (id) ON DELETE CASCADE ON UPDATE RESTRICT) ENGINE = InnoDB;

INSERT INTO newsposts(title, body, date, author) values ("test 1", "here's some content", "2013-07-18 13:44:22.123456", "Munsy");

CREATE USER 'guild'@'localhost' IDENTIFIED BY 'a';

GRANT ALL PRIVILEGES ON *.* TO 'guild'@'localhost';
*/
