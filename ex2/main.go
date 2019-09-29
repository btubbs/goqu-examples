package main

import (
	"fmt"

	"github.com/doug-martin/goqu"
)

// This example shows searching several text columns at once, each with a different weight.

func main() {
	query := buildSearchQuery("asian")
	sql, _, _ := query.ToSQL()
	fmt.Println(sql)
}

func buildSearchQuery(searchTerm string) *goqu.SelectDataset {
	/*
		SELECT name, ts_rank_cd(
			setweight(to_tsvector(name), 'A') ||
			setweight(to_tsvector(description), 'B') ||
			setweight(to_tsvector(facets::text), 'C'),
			plainto_tsquery('asian')
		) score FROM businesses
		ORDER BY score DESC;
	*/
	tsvector := goqu.L(`
		setweight(to_tsvector(name), 'A') ||
		setweight(to_tsvector(coalesce(description, '')), 'B') ||
		setweight(to_tsvector(facets::text), 'C')
	`)
	query := goqu.From("businesses").Select(
		goqu.I("name"),
		goqu.Func("ts_rank_cd",
			tsvector,
			goqu.Func("plainto_tsquery", searchTerm),
		).As("score"),
	).Order(goqu.I("score").Desc())
	return query
}
