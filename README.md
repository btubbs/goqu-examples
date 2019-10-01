# goqu-examples

This repository has a series of examples of using
[Goqu](https://github.com/doug-martin/goqu) to build progressively complex
queries that exercise Postgres's text search, JSON, and geo-spatial features.
It's the Goqu equivalent of the SQLAlchemy examples in [this blog
post](http://btubbs.com/postgres-search-with-facets-and-location-awareness.html).

## Prerequisites

Before you can run these examples, you should create a local Postgres database named "goqusearch", and run this SQL in it:

```
BEGIN;
CREATE TABLE businesses (
  id SERIAL NOT NULL, 
  name TEXT NOT NULL, 
  description TEXT, 
  street_address TEXT[] NOT NULL, 
  city TEXT NOT NULL, 
  state TEXT NOT NULL, 
  postcode TEXT NOT NULL, 
  latitude FLOAT NOT NULL, 
  longitude FLOAT NOT NULL, 
  facets JSONB DEFAULT '{}' NOT NULL, 
  PRIMARY KEY (id)
);

INSERT INTO businesses VALUES (1, 'Asian Star', 'Chinese classics plus sushi rolls from a dedicated bar, all served in an airy, contemporary space.', '{"7588 Union Park Ave"}', 'Midvale', 'UT', '84047', 40.6134920000000008, -111.857440999999994, '{"food_types": ["asian", "chinese", "japanese"], "price_range": 2, "kid_friendly": true}');
INSERT INTO businesses VALUES (2, 'Saigon Sandwich', 'a bright and stylish kitchen serving Banh Mi, Pho, and other authentic Vietnamese rice and noodle favorites.', '{"8528 S 1300 E"}', 'Sandy', 'UT', '84094', 40.5967240000000018, -111.854409000000004, '{"food_types": ["asian", "vietnamese"], "price_range": 1, "kid_friendly": true}');
INSERT INTO businesses VALUES (3, 'La Caille', 'Upscale restaurant & event venue serving French-Belgian fare in an elegant, château-style setting.', '{"9565 Wasatch Blvd"}', 'Sandy', 'UT', '84092', 40.5759040000000013, -111.792124999999999, '{"food_types": ["french", "belgian"], "price_range": 3, "kid_friendly": false}');
INSERT INTO businesses VALUES (4, 'Mezquite Mexican Grill', NULL, '{"265 7200 S"}', 'Midvale', 'UT', '84047', 40.6204999999999998, -111.898919000000006, '{"food_types": ["mexican", "tacos", "burritos"], "price_range": 1, "kid_friendly": false}');
INSERT INTO businesses VALUES (5, 'Macy''s', 'Department store chain providing brand-name clothing, accessories, home furnishings & housewares.', '{"10600 S 110 W"}', 'Sandy', 'UT', '84070', 40.6364260000000002, -111.885600999999994, '{}');
INSERT INTO businesses VALUES (6, 'Target', 'Retail chain offering home goods, clothing, electronics & more, plus exclusive designer collections.', '{"7025 S Park Centre Dr"}', 'Salt Lake City', 'UT', '84121', 40.630391000000003, -111.849552000000003, '{}');
INSERT INTO businesses VALUES (7, 'Big O Tires', NULL, '{"2284 E Fort Union Blvd"}', 'Salt Lake City', 'UT', '84121', 40.6257689999999982, -111.825850000000003, '{"services": ["tires"]}');
INSERT INTO businesses VALUES (8, 'Master Muffler and Brake Complete Auto Care', NULL, '{"6790 State St"}', 'Murray', 'UT', '84107', 40.6301330000000007, -111.890652000000003, '{"services": ["mufflers", "brakes"]}');

-- fast forward our ID sequence to account for the rows inserted with explicit IDs
SELECT pg_catalog.setval('businesses_id_seq', 8, true);

COMMIT;
```

You'll also need to enable the "earthdistance" extension that comes with Postgres.

```
CREATE EXTENSION earthdistance CASCADE;
```

Lastly, you'll need to install these Go libraries:

```
go get -u github.com/doug-martin/goqu
go get -u github.com/btubbs/pqjson
go get -u github.com/davecgh/go-spew/spew
```
