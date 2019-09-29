package main

import (
	"fmt"

	"github.com/doug-martin/goqu"
)

// In this example we do a Postgres full-text search of each business's "name" column, and rank them
// by their similarity scores.

func main() {
	query := buildSearchQuery("asian")
	sql, _, _ := query.ToSQL()
	fmt.Println(sql)
}

func buildSearchQuery(searchTerm string) *goqu.SelectDataset {
	/*
		SELECT name, ts_rank_cd(
		  to_tsvector(name),
		  plainto_tsquery('asian')
		) score FROM businesses
		ORDER BY score DESC;
	*/
	query := goqu.From("businesses").Select(
		goqu.I("name"),
		goqu.Func("ts_rank_cd",
			goqu.Func("to_tsvector", goqu.I("name")),
			goqu.Func("plainto_tsquery", searchTerm),
		).As("score"),
	).Order(goqu.I("score").Desc())
	return query
}
