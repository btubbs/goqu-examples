package main

import (
	"database/sql"
	"encoding/json"

	"github.com/btubbs/pqjson"
	"github.com/davecgh/go-spew/spew"
	"github.com/doug-martin/goqu"
	"github.com/guregu/null"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// This example shows how to use the new fancy query builder in a real app with a real Postgres
// instance, scanning the results into a Go struct.

func main() {
	db, err := sql.Open("postgres", "postgres://postgres@/goqusearch?sslmode=disable")
	check(err)
	gdb := goqu.New("postgres", db)
	businesses, err := searchBusinesses(
		gdb,
		"asian",
		FacetEq("kid_friendly", true),
		FacetLt("price_range", 2),
		Within(3, 40.606536, -111.854952),
	)
	check(err)
	spew.Dump(businesses)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func searchBusinesses(db *goqu.Database, searchTerm string, filters ...BusinessSearchFilter) ([]BusinessSearchResult, error) {
	// build the query.
	query, err := buildSearchQuery(db, searchTerm, filters...)
	if err != nil {
		return nil, err
	}
	var results []BusinessSearchResult
	// run the query in the DB and scan the results into a slice of structs.
	if err := query.ScanStructs(&results); err != nil {
		return nil, err
	}
	return results, nil
}

type Business struct {
	ID            int              `db:"id"`
	Name          string           `db:"name"`
	Description   null.String      `db:"description"`
	StreetAddress pq.StringArray   `db:"street_address"`
	City          string           `db:"city"`
	State         string           `db:"state"`
	Postcode      string           `db:"postcode"`
	Latitude      float64          `db:"latitude"`
	Longitude     float64          `db:"longitude"`
	Facets        pqjson.StringMap `db:"facets"`
}

type BusinessSearchResult struct {
	Business
	Score float64 `db:"score"`
}

func buildSearchQuery(db *goqu.Database, searchTerm string, filters ...BusinessSearchFilter) (*goqu.SelectDataset, error) {
	tsvector := goqu.L(`
		setweight(to_tsvector(name), 'A') ||
		setweight(to_tsvector(coalesce(description, '')), 'B') ||
		setweight(to_tsvector(facets::text), 'C')
	`)

	score := goqu.Func("ts_rank_cd",
		tsvector,
		goqu.Func("plainto_tsquery", searchTerm),
	)

	query := db.From("businesses").Select(
		Business{},        // grab all the columns defined on the Business struct
		score.As("score"), // plus our computed "score" column
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
			goqu.Cast(goqu.Cast(facet, "TEXT"), "NUMERIC").Lt(value),
		), nil
	}
}

func Within(miles, lat, long float64) BusinessSearchFilter {
	return func(query *goqu.SelectDataset) (*goqu.SelectDataset, error) {
		// NOTE that the Postgres "point" function takes longitude first and latitude second, which you
		// might not expect.
		bizPoint := goqu.Func("point", goqu.I("longitude"), goqu.I("latitude"))
		myPoint := goqu.Func("point", long, lat)
		distance := goqu.L("? <@> ?", bizPoint, myPoint)
		return query.Where(distance.Lt(miles)), nil
	}
}
