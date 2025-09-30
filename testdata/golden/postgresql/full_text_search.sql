-- Full text search queries with various configurations
select
  id,
  title,
  content,
  ts_rank(document, plainto_tsquery('english', 'database performance')) as relevance_rank,
  ts_rank_cd(document, plainto_tsquery('english', 'database performance')) as relevance_rank_cd,
  ts_headline('english', content, plainto_tsquery('english', 'database performance')) as highlighted_content,
  document @@ plainto_tsquery('english', 'database performance') as matches_search,
  ts_rank(document, to_tsquery('english', 'database & performance')) as exact_rank,
  ts_rank(document, to_tsquery('english', 'database | performance')) as or_rank
from
  articles
where
  document @@ plainto_tsquery('english', 'database performance')
order by
  ts_rank(document, plainto_tsquery('english', 'database performance')) desc;

-- Advanced full text search with multiple languages and configurations
select
  product_id,
  product_name,
  description,
  language,
  to_tsvector(language::regconfig, description) as document_vector,
  to_tsquery(language::regconfig, 'high quality') as quality_query,
  ts_rank(to_tsvector(language::regconfig, description), to_tsquery(language::regconfig, 'high quality')) as quality_rank,
  ts_headline(language::regconfig, description, to_tsquery(language::regconfig, 'high quality')) as highlighted_desc,
  length(description) as desc_length,
  array_length(tsvector_to_array(to_tsvector(language::regconfig, description)), 1) as unique_words
from
  products
where
  to_tsvector(language::regconfig, description) @@ to_tsquery(language::regconfig, 'high quality')
order by
  quality_rank desc,
  language;

-- Full text search with phrase search and distance operators
select
  id,
  title,
  content,
  ts_rank(document, phraseto_tsquery('english', 'machine learning')) as phrase_rank,
  ts_rank(document, to_tsquery('english', 'machine <-> learning')) as proximity_rank,
  ts_headline('english', content, phraseto_tsquery('english', 'machine learning')) as phrase_highlight,
  document @@ phraseto_tsquery('english', 'machine learning') as exact_phrase_match,
  document @@ to_tsquery('english', 'machine <2> learning') as close_proximity_match,
  document @@ to_tsquery('english', 'machine <-> learning <-> algorithms') as phrase_sequence_match,
  ts_rank(document, to_tsquery('english', 'machine <1> learning')) as adjacent_words_rank
from
  research_papers
where
  document @@ phraseto_tsquery('english', 'machine learning')
order by
  phrase_rank desc;

-- Full text search with weights and multiple columns
select
  id,
  title,
  content,
  keywords,
  setweight(to_tsvector('english', title), 'A') || setweight(to_tsvector('english', content), 'B') || setweight(to_tsvector('english', keywords), 'C') as weighted_document,
  ts_rank(
    setweight(to_tsvector('english', title), 'A') || setweight(to_tsvector('english', content), 'B') || setweight(to_tsvector('english', keywords), 'C'),
    plainto_tsquery('english', 'data science')
  ) as weighted_rank,
  ts_headline('english', title, plainto_tsquery('english', 'data science')) as title_highlight,
  ts_headline('english', content, plainto_tsquery('english', 'data science')) as content_highlight,
  case
    when ts_rank(to_tsvector('english', title), plainto_tsquery('english', 'data science')) > 0.5 then 'title_match'
    when ts_rank(to_tsvector('english', content), plainto_tsquery('english', 'data science')) > 0.3 then 'content_match'
    else 'keyword_match'
  end as match_type
from
  blog_posts
where
  setweight(to_tsvector('english', title), 'A') || setweight(to_tsvector('english', content), 'B') || setweight(to_tsvector('english', keywords), 'C') @@ plainto_tsquery('english', 'data science')
order by
  weighted_rank desc;

-- Full text search with custom dictionaries and configurations
select
  id,
  title,
  content,
  ts_rank(document, plainto_tsquery('english', 'artificial intelligence')) as ai_rank,
  ts_rank(document, plainto_tsquery('simple', 'artificial intelligence')) as simple_ai_rank,
  ts_debug('english', content) as english_analysis,
  ts_debug('simple', content) as simple_analysis,
  ts_lexize('english_stem', 'running') as stemmed_running,
  ts_lexize('english_stem', 'ran') as stemmed_ran,
  ts_lexize('english_stem', 'runs') as stemmed_runs,
  length(content) as content_length,
  array_length(tsvector_to_array(document), 1) as unique_terms
from
  documents
where
  document @@ plainto_tsquery('english', 'artificial intelligence')
order by
  ai_rank desc;

-- Full text search with aggregation and statistics
select
  search_term,
  count(*) as search_count,
  avg(ts_rank(document, plainto_tsquery('english', search_term))) as avg_relevance,
  max(ts_rank(document, plainto_tsquery('english', search_term))) as max_relevance,
  min(ts_rank(document, plainto_tsquery('english', search_term))) as min_relevance,
  percentile_cont(0.5) within group(
    order by
      ts_rank(document, plainto_tsquery('english', search_term))
  ) as median_relevance,
  string_agg(distinct title, '; ') as matching_titles
from
  (
    select
      unnest(
        array ['machine learning', 'data science', 'artificial intelligence', 'neural networks']
      ) as search_term
  ) terms
  cross join articles
where
  document @@ plainto_tsquery('english', search_term)
group by
  search_term
order by
  search_count desc,
  avg_relevance desc;

-- Full text search with triggers and indexing
create index concurrently if not exists articles_document_idx on articles using gin(document);

create or replace function
  articles_tsvector_trigger() returns trigger as $tsvector_trigger$ begin if tg_op = 'INSERT' then new.document = setweight(to_tsvector('english', new.title), 'A') || setweight(to_tsvector('english', new.content), 'B') || setweight(to_tsvector('english', new.tags), 'C');
return new;
elsif tg_op = 'UPDATE' then if new.title != old.title
or new.content != old.content
or new.tags != old.tags then new.document = setweight(to_tsvector('english', new.title), 'A') || setweight(to_tsvector('english', new.content), 'B') || setweight(to_tsvector('english', new.tags), 'C');
end if;
return new;
else return old;
end if;
end $tsvector_trigger$ language plpgsql;

create trigger articles_tsvector_update before insert
or
update on articles for each row execute function articles_tsvector_trigger();

-- Query using the indexed full text search
select
  id,
  title,
  ts_rank(document, plainto_tsquery('english', 'quantum computing')) as relevance,
  ts_headline('english', content, plainto_tsquery('english', 'quantum computing')) as highlighted_content
from
  articles
where
  document @@ plainto_tsquery('english', 'quantum computing')
order by
  relevance desc
limit
  10;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/postgresql/full_text_search.sql