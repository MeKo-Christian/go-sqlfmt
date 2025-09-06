select name,email,type from`travel-sample`where type='airline'and country='United States'order by name;

select h.name as hotel_name,r.author,r.content,r.ratings.Overall as overall_rating from`travel-sample`h unnest h.reviews r where h.type='hotel'and h.country='United Kingdom'and r.ratings.Overall>=4 order by r.ratings.Overall desc;