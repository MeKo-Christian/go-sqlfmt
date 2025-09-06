select
  u.username,
  p.title as post_title,
  c.content as comment_content,
  count(l.id) as like_count,
  avg(r.rating) as avg_rating
from
  users u
  inner join posts p on u.id = p.author_id
  left join comments c on p.id = c.post_id
  left join likes l on p.id = l.post_id
  left join ratings r on p.id = r.post_id
where
  u.active = true
  and p.published = true
  and p.created_at >= '2023-01-01'
  and (
    c.approved = true
    or c.id is null
  )
group by
  u.id,
  u.username,
  p.id,
  p.title,
  c.id,
  c.content
having
  count(l.id) > 5
order by
  like_count desc,
  avg_rating desc
limit
  50;