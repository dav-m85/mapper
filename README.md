# mapper
mapper provides a way of mapping a `sql/database` Scan-like output
to a struct, and some helper functions to write them.

When manipulating raw queries, it's easy to introduce bug by misordering fields,
forgetting about names.

```
var ma = Mapper(Record{}, "*")

db := sql.DB{}

rows, _ := db.Query(`SELECT ` + ma.ColumnsString() + ` FROM Records`)
for rows.Next() {
    var r Record
    rows.Scan(ma.Addrs(&r)...)
}
```

## Installation
```
go get github.com/dav-m85/mapper
```

Or you can just copypaste mapper.go in your project directly. Less deps is good.
