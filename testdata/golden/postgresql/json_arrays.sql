select
  id,
  data ->> 'name' as name,
  data -> 'profile' ->> 'email' as email,
  jsonb_array_length(data -> 'tags') as tag_count
from
  users
where
  data?'active'and(data ->> 'active')::boolean = true
order by
(data ->> 'created_at')::timestamp desc;

select
  name,
  tags,
  array_length(tags, 1) as tag_count
from
  articles
where
  'postgresql' = any(tags)
  and 'database' = any(tags)
order by
  array_length(tags, 1) desc;