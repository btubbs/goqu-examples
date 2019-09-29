package main

import (
	"fmt"

	"github.com/doug-martin/goqu"
)

// This example filters down the results to only include those with a score greater than 0 It shows
// where using a query builder is starting to get less verbose than the equivalent raw SQL, since we
// can re-use the "score" clause in both the Select and the Where.  In SQL you'd have to repeat the
// whole thing or stick it in a CTE.

func main() {
	query := buildSearchQuery("asian")
	sql, _, _ := query.ToSQL()
	fmt.Println(sql)
}

func buildSearchQuery(searchTerm string) *goqu.SelectDataset {
	tsvector := goqu.L(`
		setweight(to_tsvector(name), 'A') ||
		setweight(to_tsvector(coalesce(description, '')), 'B') ||
		setweight(to_tsvector(facets::text), 'C')
	`)

	score := goqu.Func("ts_rank_cd",
		tsvector,
		goqu.Func("plainto_tsquery", searchTerm),
	)

	query := goqu.From("businesses").Select(
		goqu.I("name"),
		score.As("score"),
	).Order(goqu.I("score").Desc()).Where(score.Gt(0))
	return query
}
