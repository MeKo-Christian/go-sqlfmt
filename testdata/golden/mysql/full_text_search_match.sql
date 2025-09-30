-- MySQL Full Text Search with MATCH AGAINST
SELECT
    id,
    title,
    content,
    MATCH(title, content) AGAINST('database performance' IN NATURAL LANGUAGE MODE) AS relevance_score,
    MATCH(title, content) AGAINST('database performance' IN BOOLEAN MODE) AS boolean_score,
    MATCH(title, content) AGAINST('database performance' WITH QUERY EXPANSION) AS expansion_score
FROM articles
WHERE MATCH(title, content) AGAINST('database performance' IN NATURAL LANGUAGE MODE)
ORDER BY relevance_score DESC;

-- Full text search with multiple columns and weights
SELECT
    product_id,
    product_name,
    description,
    specifications,
    MATCH(product_name, description) AGAINST('high quality' IN NATURAL LANGUAGE MODE) AS name_desc_score,
    MATCH(specifications) AGAINST('durable waterproof' IN BOOLEAN MODE) AS specs_score,
    (
        MATCH(product_name, description) AGAINST('high quality' IN NATURAL LANGUAGE MODE) * 2 +
        MATCH(specifications) AGAINST('durable waterproof' IN BOOLEAN MODE)
    ) / 3 AS weighted_score
FROM products
WHERE MATCH(product_name, description, specifications) AGAINST('high quality durable' IN BOOLEAN MODE)
ORDER BY weighted_score DESC;

-- Boolean mode full text search with operators
SELECT
    id,
    title,
    content,
    MATCH(title, content) AGAINST('+database -mysql +postgresql' IN BOOLEAN MODE) AS boolean_score
FROM technical_docs
WHERE MATCH(title, content) AGAINST('+database +performance +(optimization OR tuning)' IN BOOLEAN MODE)
ORDER BY boolean_score DESC;

-- Full text search with query expansion
SELECT
    article_id,
    title,
    summary,
    MATCH(title, summary, content) AGAINST('machine learning' WITH QUERY EXPANSION) AS expanded_score,
    MATCH(title, summary, content) AGAINST('machine learning' IN NATURAL LANGUAGE MODE) AS natural_score
FROM articles
WHERE MATCH(title, summary, content) AGAINST('AI neural networks' WITH QUERY EXPANSION)
ORDER BY expanded_score DESC
LIMIT 20;

-- Full text search with relevance and ranking
SELECT
    id,
    title,
    author,
    publication_date,
    MATCH(title, abstract, keywords) AGAINST('quantum computing' IN NATURAL LANGUAGE MODE) AS relevance,
    CASE
        WHEN MATCH(title, abstract, keywords) AGAINST('quantum computing' IN NATURAL LANGUAGE MODE) > 0.8 THEN 'highly_relevant'
        WHEN MATCH(title, abstract, keywords) AGAINST('quantum computing' IN NATURAL LANGUAGE MODE) > 0.5 THEN 'relevant'
        ELSE 'somewhat_relevant'
    END AS relevance_category,
    DATEDIFF(CURDATE(), publication_date) AS days_since_publication
FROM research_papers
WHERE MATCH(title, abstract, keywords) AGAINST('quantum computing' IN NATURAL LANGUAGE MODE)
ORDER BY relevance DESC, days_since_publication ASC;

-- Full text search with phrase matching and wildcards
SELECT
    product_id,
    product_name,
    description,
    MATCH(product_name, description) AGAINST('"wireless headphones" bluetooth*' IN BOOLEAN MODE) AS phrase_score
FROM products
WHERE MATCH(product_name, description) AGAINST('"wireless headphones" +bluetooth' IN BOOLEAN MODE)
ORDER BY phrase_score DESC;

-- Full text search with minimum word length and stopwords
SELECT
    id,
    title,
    content,
    MATCH(title, content) AGAINST('web development javascript' IN NATURAL LANGUAGE MODE) AS score,
    LENGTH(content) AS content_length
FROM blog_posts
WHERE MATCH(title, content) AGAINST('web development javascript' IN NATURAL LANGUAGE MODE)
    AND LENGTH(content) > 1000
ORDER BY score DESC;

-- Full text search with aggregation and statistics
SELECT
    search_term,
    COUNT(*) AS matches_found,
    AVG(MATCH(title, content) AGAINST(search_term IN NATURAL LANGUAGE MODE)) AS avg_relevance,
    MAX(MATCH(title, content) AGAINST(search_term IN NATURAL LANGUAGE MODE)) AS max_relevance,
    MIN(MATCH(title, content) AGAINST(search_term IN NATURAL LANGUAGE MODE)) AS min_relevance,
    STDDEV(MATCH(title, content) AGAINST(search_term IN NATURAL LANGUAGE MODE)) AS relevance_stddev
FROM (
    SELECT 'database' AS search_term
    UNION ALL SELECT 'performance'
    UNION ALL SELECT 'optimization'
    UNION ALL SELECT 'indexing'
) terms
CROSS JOIN articles
WHERE MATCH(title, content) AGAINST(search_term IN NATURAL LANGUAGE MODE)
GROUP BY search_term
ORDER BY matches_found DESC;

-- Full text search with user-generated content and filtering
SELECT
    review_id,
    product_id,
    user_id,
    rating,
    review_text,
    MATCH(review_text) AGAINST('excellent amazing wonderful' IN NATURAL LANGUAGE MODE) AS positive_score,
    MATCH(review_text) AGAINST('terrible awful horrible' IN NATURAL LANGUAGE MODE) AS negative_score,
    CASE
        WHEN MATCH(review_text) AGAINST('excellent amazing wonderful' IN NATURAL LANGUAGE MODE) >
             MATCH(review_text) AGAINST('terrible awful horrible' IN NATURAL LANGUAGE MODE) THEN 'positive'
        WHEN MATCH(review_text) AGAINST('terrible awful horrible' IN NATURAL LANGUAGE MODE) >
             MATCH(review_text) AGAINST('excellent amazing wonderful' IN NATURAL LANGUAGE MODE) THEN 'negative'
        ELSE 'neutral'
    END AS sentiment
FROM product_reviews
WHERE MATCH(review_text) AGAINST('quality performance reliability' IN BOOLEAN MODE)
    AND rating >= 3
ORDER BY positive_score DESC;

-- Full text search with custom relevancy scoring
SELECT
    id,
    title,
    content,
    author,
    view_count,
    MATCH(title, content) AGAINST('artificial intelligence' IN NATURAL LANGUAGE MODE) AS base_relevance,
    LOG10(view_count + 1) AS popularity_score,
    DATEDIFF(CURDATE(), created_at) AS days_old,
    (
        MATCH(title, content) AGAINST('artificial intelligence' IN NATURAL LANGUAGE MODE) * 0.7 +
        LOG10(view_count + 1) * 0.2 +
        (1 / LOG10(DATEDIFF(CURDATE(), created_at) + 2)) * 0.1
    ) AS custom_score
FROM articles
WHERE MATCH(title, content) AGAINST('artificial intelligence' IN NATURAL LANGUAGE MODE)
ORDER BY custom_score DESC;

-- Full text search with multiple indexes and joins
SELECT
    p.id,
    p.name AS product_name,
    c.name AS category_name,
    MATCH(p.name, p.description) AGAINST('gaming laptop' IN NATURAL LANGUAGE MODE) AS product_score,
    MATCH(c.description) AGAINST('electronics computers' IN NATURAL LANGUAGE MODE) AS category_score,
    (MATCH(p.name, p.description) AGAINST('gaming laptop' IN NATURAL LANGUAGE MODE) +
     MATCH(c.description) AGAINST('electronics computers' IN NATURAL LANGUAGE MODE)) / 2 AS combined_score
FROM products p
JOIN categories c ON p.category_id = c.id
WHERE MATCH(p.name, p.description) AGAINST('gaming laptop' IN NATURAL LANGUAGE MODE)
   OR MATCH(c.description) AGAINST('electronics computers' IN NATURAL LANGUAGE MODE)
ORDER BY combined_score DESC;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/mysql/full_text_search_match.sql