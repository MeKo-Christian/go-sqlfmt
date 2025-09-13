# Couchbase N1QL Dialect

N1QL (SQL for JSON) dialect provides formatting for Couchbase N1QL queries designed for JSON document databases.

## Basic Usage

### Library Usage

Use the N1QL dialect by setting the language to `sqlfmt.N1QL`:

```go
cfg := sqlfmt.NewDefaultConfig().WithLang(sqlfmt.N1QL)
fmt.Println(sqlfmt.Format(query, cfg))
```

### CLI Usage

```bash
# Format N1QL queries
sqlfmt format --lang=n1ql query.n1ql
```

## Supported Features

### N1QL-Specific Query Features

#### FROM Clause with Buckets

```sql
SELECT name, email, type
FROM `travel-sample`
WHERE type = "airline";
```

#### USE KEYS Clause

```sql
SELECT *
FROM `travel-sample`
USE KEYS ["airline_10", "airline_137"];
```

#### JOIN Operations with Document References

```sql
SELECT airline.name, route.sourceairport, route.destinationairport
FROM `travel-sample` airline
JOIN `travel-sample` route
  ON KEYS route.airlineid
WHERE airline.type = "airline"
  AND route.type = "route";
```

#### NEST Operations

```sql
SELECT hotel.name,
       ARRAY r FOR r IN hotel.reviews WHEN r.ratings.Overall >= 4 END AS good_reviews
FROM `travel-sample` hotel
WHERE hotel.type = "hotel"
  AND hotel.city = "San Francisco";
```

### JSON Path Expressions

#### Dot Notation

```sql
SELECT
  airline.name,
  airline.country,
  airline.icao,
  airline.callsign
FROM `travel-sample` airline
WHERE airline.type = "airline"
  AND airline.country = "United States";
```

#### Bracket Notation

```sql
SELECT
  hotel["name"],
  hotel["address"],
  hotel["city"]
FROM `travel-sample` hotel
WHERE hotel["type"] = "hotel";
```

#### Array Indexing

```sql
SELECT
  landmark.name,
  landmark.activity[0] AS primary_activity,
  landmark.geo.lat,
  landmark.geo.lon
FROM `travel-sample` landmark
WHERE landmark.type = "landmark";
```

### N1QL Functions

#### JSON Functions

```sql
SELECT
  name,
  OBJECT_LENGTH(address) AS address_fields,
  OBJECT_NAMES(geo) AS geo_properties,
  OBJECT_VALUES(reviews[0].ratings) AS first_review_ratings
FROM `travel-sample`
WHERE type = "hotel"
LIMIT 10;
```

#### Array Functions

```sql
SELECT
  name,
  ARRAY_LENGTH(reviews) AS review_count,
  ARRAY_AVG(ARRAY r.ratings.Overall FOR r IN reviews END) AS avg_rating,
  ARRAY_SUM(ARRAY r.ratings.Value FOR r IN reviews END) AS total_value_rating
FROM `travel-sample`
WHERE type = "hotel"
  AND ARRAY_LENGTH(reviews) > 0;
```

#### String Functions

```sql
SELECT
  UPPER(name) AS hotel_name,
  LOWER(city) AS city_name,
  LENGTH(description) AS desc_length,
  SUBSTR(phone, 0, 3) AS area_code
FROM `travel-sample`
WHERE type = "hotel";
```

#### Type Functions

```sql
SELECT
  name,
  TYPE(geo) AS geo_type,
  TYPENAME(address) AS address_typename,
  ISSTRING(name) AS name_is_string,
  ISNUMBER(id) AS id_is_number
FROM `travel-sample`
WHERE type = "landmark";
```

### ARRAY Operations

#### ARRAY Constructor

```sql
SELECT
  name,
  ARRAY activity FOR activity IN activities WHEN activity != "golf" END AS non_golf_activities
FROM `travel-sample`
WHERE type = "hotel";
```

#### ARRAY Comprehensions

```sql
SELECT
  name,
  ARRAY {
    "reviewer": r.author,
    "rating": r.ratings.Overall,
    "date": r.date
  } FOR r IN reviews WHEN r.ratings.Overall >= 4 END AS good_reviews
FROM `travel-sample`
WHERE type = "hotel";
```

#### ANY and EVERY

```sql
-- ANY: At least one element satisfies condition
SELECT name
FROM `travel-sample`
WHERE type = "hotel"
  AND ANY r IN reviews SATISFIES r.ratings.Overall >= 5 END;

-- EVERY: All elements satisfy condition
SELECT name
FROM `travel-sample`
WHERE type = "hotel"
  AND EVERY r IN reviews SATISFIES r.ratings.Overall >= 3 END;
```

### Subqueries and Common Table Expressions

#### Subqueries

```sql
SELECT h.name, h.city, h.country
FROM `travel-sample` h
WHERE h.type = "hotel"
  AND h.city IN (
    SELECT DISTINCT l.city
    FROM `travel-sample` l
    WHERE l.type = "landmark"
      AND l.country = "United States"
  );
```

#### WITH Clause (Common Table Expressions)

```sql
WITH top_airlines AS (
  SELECT a.name, a.country, COUNT(r) AS route_count
  FROM `travel-sample` a
  JOIN `travel-sample` r ON KEYS r.airlineid
  WHERE a.type = "airline" AND r.type = "route"
  GROUP BY a.name, a.country
  HAVING COUNT(r) > 100
),
us_airports AS (
  SELECT ap.airportname, ap.city, ap.faa
  FROM `travel-sample` ap
  WHERE ap.type = "airport" AND ap.country = "United States"
)
SELECT ta.name AS airline, ua.airportname, ua.city
FROM top_airlines ta, us_airports ua
WHERE ta.country = "United States"
ORDER BY ta.route_count DESC, ua.city;
```

### N1QL Index Hints

#### USE INDEX

```sql
SELECT name, city, country
FROM `travel-sample` USE INDEX (`def_type`)
WHERE type = "hotel"
  AND country = "United States";
```

#### AVOID INDEX

```sql
SELECT *
FROM `travel-sample` AVOID INDEX (`def_name_type`)
WHERE name LIKE "Marriott%";
```

### Aggregation and Grouping

#### GROUP BY with JSON Paths

```sql
SELECT
  country,
  COUNT(*) AS hotel_count,
  AVG(ARRAY_AVG(ARRAY r.ratings.Overall FOR r IN reviews END)) AS avg_rating
FROM `travel-sample`
WHERE type = "hotel"
  AND ARRAY_LENGTH(reviews) > 0
GROUP BY country
HAVING COUNT(*) > 10
ORDER BY avg_rating DESC;
```

#### Window Functions

```sql
SELECT
  name,
  city,
  country,
  ARRAY_AVG(ARRAY r.ratings.Overall FOR r IN reviews END) AS avg_rating,
  ROW_NUMBER() OVER (
    PARTITION BY country
    ORDER BY ARRAY_AVG(ARRAY r.ratings.Overall FOR r IN reviews END) DESC
  ) AS country_rank
FROM `travel-sample`
WHERE type = "hotel"
  AND ARRAY_LENGTH(reviews) > 0;
```

### UPSERT and MERGE Operations

#### UPSERT

```sql
UPSERT INTO `travel-sample` (KEY, VALUE)
VALUES ("airline_new", {
  "type": "airline",
  "id": 9999,
  "name": "New Airline",
  "iata": "NA",
  "icao": "NEW",
  "callsign": "NEWAIR",
  "country": "United States"
});
```

#### MERGE

```sql
MERGE INTO `travel-sample` t
USING [
  {
    "id": 1001,
    "type": "airline",
    "name": "Updated Airline",
    "country": "Canada"
  }
] s ON KEY "airline_" || s.id
WHEN MATCHED THEN
  UPDATE SET t.name = s.name, t.country = s.country
WHEN NOT MATCHED THEN
  INSERT {
    "type": s.type,
    "id": s.id,
    "name": s.name,
    "country": s.country
  };
```

### Data Modification

#### INSERT

```sql
INSERT INTO `travel-sample` (KEY, VALUE)
VALUES ("landmark_new", {
  "type": "landmark",
  "name": "Golden Gate Bridge",
  "city": "San Francisco",
  "state": "California",
  "country": "United States",
  "geo": {
    "lat": 37.8199,
    "lon": -122.4783
  },
  "activity": ["photography", "sightseeing"]
});
```

#### UPDATE

```sql
UPDATE `travel-sample`
SET address.country = "USA"
WHERE type = "hotel"
  AND address.country = "United States";
```

#### DELETE

```sql
DELETE FROM `travel-sample`
WHERE type = "airline"
  AND country IS MISSING;
```

## N1QL-Specific Features

### MISSING vs NULL

```sql
SELECT name
FROM `travel-sample`
WHERE type = "hotel"
  AND (phone IS MISSING OR phone IS NULL);
```

### Meta() Function

```sql
SELECT
  META().id,
  META().type,
  name,
  city
FROM `travel-sample`
WHERE type = "hotel"
LIMIT 5;
```

### Document Expiration

```sql
SELECT
  name,
  META().expiration
FROM `travel-sample`
WHERE type = "hotel"
  AND META().expiration > 0;
```

## Current Limitations

### Complex JSON Operations

- Very complex nested JSON manipulations may need manual formatting
- Some advanced N1QL functions may not have specialized formatting rules

### Index Optimization

- Index hints are preserved but not validated
- Query optimization suggestions are not provided

## Testing

### Run N1QL Tests

```bash
# All N1QL tests
go test ./pkg/sqlfmt -run TestN1QL

# Golden file tests
just test-golden
```

### Test Data Locations

- **Input files**: `testdata/input/n1ql/*.n1ql`
- **Expected output**: `testdata/golden/n1ql/*.n1ql`

## Implementation Status

**Current Status**: ✅ **Comprehensive N1QL support implemented**

- [x] JSON path expressions and navigation
- [x] N1QL-specific functions (JSON, Array, Type functions)
- [x] ARRAY operations and comprehensions
- [x] ANY and EVERY operators
- [x] USE KEYS and JOIN operations
- [x] NEST operations
- [x] Index hints (USE INDEX, AVOID INDEX)
- [x] UPSERT and MERGE operations
- [x] Document modification (INSERT, UPDATE, DELETE)
- [x] Meta functions and document properties
- [x] Window functions for JSON documents

The N1QL dialect provides comprehensive support for Couchbase's SQL-for-JSON query language, enabling proper formatting of complex document-oriented queries and operations.
