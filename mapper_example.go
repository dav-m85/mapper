package mapper

import (
	"database/sql"
	"time"
)

type Record struct {
	ID   string
	Name string
}

func ExampleSimple() {
	var ma = Mapper(Record{}, "*")

	db := sql.DB{}

	rows, _ := db.Query(`SELECT ` + ma.ColumnsString() + ` FROM Records`)
	for rows.Next() {
		var r Record
		rows.Scan(ma.Addrs(&r)...)
	}
}

// ExampleUse for a table created with:
// CREATE TABLE IF NOT EXISTS activities (
//
//	id INTEGER NOT NULL PRIMARY KEY,
//	time DATETIME NOT NULL,
//	description TEXT
func ExampleMapper() {
	var db *sql.DB

	type Activity struct {
		ID          int
		Time        time.Time
		Description string
	}
	a := Activity{
		ID:          1,
		Time:        time.Now(),
		Description: "an activity",
	}
	mapper := Mapper(Activity{}, "*")
	db.Exec(`INSERT INTO activities VALUES(`+mapper.Marks()+`);`, mapper.Values(a)...)
	// like db.Exec(`INSERT INTO activities VALUES(id, time, description);`, a.ID, a.Time, a.Description)

	b := new(Activity)

	row, _ := db.Query("SELECT * FROM activities WHERE id=?", 1)
	row.Scan(mapper.Addrs(b)...)
	// like row.Scan(&b.ID, &b.Time, &b.Description)

	pmapper := Mapper(Activity{}, "time")

	c := new(Activity)

	rowp, _ := db.Query(`SELECT `+pmapper.ColumnsString()+` FROM activities WHERE id=?`, 1)
	// db.Query(`SELECT time FROM activities WHERE id=?`, 1)
	rowp.Scan(mapper.Addrs(c)...)
	// like row.Scan(&c.Time)
}
