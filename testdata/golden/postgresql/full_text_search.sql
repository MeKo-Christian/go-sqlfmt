select
  id,
  title,
  content,
  ts_rank(
    document,
    plainto_tsquery('english', 'database performance')
  ) as relevance_rank,
  ts_rank_cd(
    document,
    plainto_tsquery('english', 'database performance')
  ) as relevance_rank_cd,
  ts_headline(
    'english',
    content,
    plainto_tsquery('english', 'database performance')
  ) as highlighted_content,
  document@@plainto_tsquery('english', 'database performance') as matches_search,
  ts_rank(
    document,
    to_tsquery('english', 'database & performance')
  ) as exact_rank,
  ts_rank(
    document,
    to_tsquery('english', 'database | performance')
  ) as or_rank
from
  articles
where
  document@@plainto_tsquery('english', 'database performance')
order by
  ts_rank(
    document,
    plainto_tsquery('english', 'database performance')
  ) desc;