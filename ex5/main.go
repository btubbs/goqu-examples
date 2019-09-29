package main

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu"
)

// Here we add filtering on key/value pairs in the JSONB column of the businesses table.  Notice the
// introduction of the "functional options" pattern in the buildSearchQuery arguments.

func main() {
	query, err := buildSearchQuery(
		"asian",
		FacetLt("price_range", 2),
	)
	if err != nil {
		panic(err)
	}
	sql, _, _ := query.ToSQL()
	fmt.Println(sql)
}

func buildSearchQuery(searchTerm string, filters ...BusinessSearchFilter) (*goqu.SelectDataset, error) {
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

	// apply all of our filters
	for _, filter := range filters {
		var err error
		query, err = filter(query)
		if err != nil {
			return nil, err
		}
	}
	return query, nil
}

// A BusinessSearchFilter is a function that takes a query and gives you back a modified one.
type BusinessSearchFilter func(*goqu.SelectDataset) (*goqu.SelectDataset, error)

// FacetEq modifies a query to filter down to where the facets column contains specific key/value
// pairs.
func FacetEq(key string, value interface{}) BusinessSearchFilter {
	return func(query *goqu.SelectDataset) (*goqu.SelectDataset, error) {
		// make a JSON object with the key/value pair we're searching for, then ask Postgres if the
		// "facets" column contains that object.
		facetData, err := json.Marshal(map[string]interface{}{key: value})
		if err != nil {
			return nil, err
		}
		return query.Where(
			goqu.L("facets @> ?", facetData),
		), nil
	}
}

func FacetGt(key string, value float64) BusinessSearchFilter {
	return func(query *goqu.SelectDataset) (*goqu.SelectDataset, error) {
		// filter for cases where the json value is numeric and greater than the passed in value.
		facet := goqu.L("facets -> ?", key)
		return query.Where(
			goqu.Func("jsonb_typeof", facet).Eq("number"),
			goqu.Cast(facet, "NUMERIC").Gt(value),
		), nil
	}
}

func FacetLt(key string, value float64) BusinessSearchFilter {
	return func(query *goqu.SelectDataset) (*goqu.SelectDataset, error) {
		// filter for cases where the json value is numeric and less than the passed in value.
		facet := goqu.L("facets -> ?", key)
		return query.Where(
			goqu.Func("jsonb_typeof", facet).Eq("number"),
			goqu.Cast(facet, "NUMERIC").Lt(value),
		), nil
	}
}
