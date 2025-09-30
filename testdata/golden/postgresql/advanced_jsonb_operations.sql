-- Advanced JSON/JSONB operations with complex queries
select
  id,
  name,
  metadata -> 'profile' ->> 'email' as email,
  metadata -> 'profile' -> 'social' ->> 'twitter' as twitter_handle,
  metadata -> 'skills' as skills_array,
  jsonb_array_length(metadata -> 'skills') as skill_count,
  metadata -> 'experience' ->> 'years' as years_experience,
(metadata -> 'experience' ->> 'years')::integer as years_experience_int,
  metadata #> '{profile,social,twitter}' as twitter_path,
  metadata #>> '{profile,social,twitter}' as twitter_path_text,
  metadata -> 'projects' -> 0 ->> 'name' as first_project_name,
  jsonb_typeof(metadata -> 'skills') as skills_type,
  jsonb_typeof(metadata -> 'profile') as profile_type,
  metadata ? 'skills' as has_skills,
  metadata ?| array ['skills', 'experience'] as has_skills_or_experience,
  metadata ?& array ['profile', 'skills'] as has_profile_and_skills,
  metadata -> 'tags' ? 'urgent' as has_urgent_tag,
  metadata -> 'tags' ?| array ['urgent', 'important'] as has_urgent_or_important,
  jsonb_object_keys(metadata -> 'profile') as profile_keys
from
  users
where
  metadata -> 'profile' ->> 'active' = 'true'
  and jsonb_array_length(metadata -> 'skills') > 2
order by
(metadata -> 'experience' ->> 'years')::integer desc;

-- JSONB containment and existence operators with complex conditions
select
  product_id,
  product_name,
  specifications,
  specifications @> '{"category": "electronics"}' as is_electronics,
  specifications @> '{"brand": "Apple"}' as is_apple_product,
  specifications <@ '{"category": "electronics", "brand": "Apple"}' as matches_criteria,
  specifications ? 'warranty' as has_warranty_info,
  specifications ?| array ['dimensions', 'weight', 'color'] as has_physical_specs,
  specifications ?& array ['price', 'currency'] as has_price_info,
  specifications #> '{specifications,battery}' ->> 'capacity' as battery_capacity,
  jsonb_extract_path_text(specifications, 'specifications', 'display', 'resolution') as display_resolution,
  specifications - 'internal_id' as specs_without_id,
  specifications - array ['created_at', 'updated_at'] as specs_clean,
  specifications || '{"verified": true}' as verified_specs
from
  products
where
  specifications @> '{"category": "electronics"}'
  and specifications -> 'price' ->> 'amount' :: numeric > 100
order by
  specifications -> 'price' ->> 'amount' :: numeric desc;

-- JSONB aggregation and array operations
select
  department,
  jsonb_agg(
    jsonb_build_object(
      'id',
      id,
      'name',
      name,
      'salary',
      salary,
      'skills',
      metadata -> 'skills'
    )
  ) as employees_json,
  jsonb_object_agg(name, salary) as name_salary_map,
  jsonb_agg(metadata -> 'skills') as all_skills_arrays,
  count(*) as employee_count,
  avg(salary) as avg_salary,
  jsonb_array_length(jsonb_agg(metadata -> 'skills')) as total_skill_arrays
from
  employees
where
  active = true
  and metadata -> 'skills' is not null
group by
  department
having
  count(*) > 1
order by
  count(*) desc;

-- JSONB set operations and transformations
select
  user_id,
  preferences,
  preferences || '{"theme": "dark"}' as dark_theme_prefs,
  preferences - 'notifications' as prefs_without_notifications,
  preferences # - '{social,auto_share}' as prefs_without_auto_share,
  jsonb_set(preferences, '{social,privacy}', '"strict"') as strict_privacy_prefs,
  jsonb_set_lax(
    preferences,
    '{notifications,email}',
    'true',
    true,
    'return_target'
  ) as email_notifications_enabled,
  jsonb_insert(preferences, '{features,0}', '"dashboard"') as dashboard_first_feature,
  preferences @> '{"theme": "light"}' as has_light_theme,
  preferences <@ '{"theme": "light", "language": "en"}' as matches_basic_prefs,
  jsonb_pretty(preferences) as formatted_preferences
from
  user_preferences
where
  preferences ? 'theme'
order by
  user_id;

-- Complex JSONB queries with path queries and array operations
select
  id,
  name,
  config,
  config -> 'database' ->> 'host' as db_host,
  config -> 'database' ->> 'port' as db_port,
  jsonb_array_elements(config -> 'features') as feature,
  jsonb_array_elements_text(config -> 'tags') as tag,
  jsonb_object_keys(config -> 'settings') as setting_key,
  config #>> '{database,credentials,password}' as db_password,
  jsonb_extract_path(config, 'monitoring', 'alerts') as alert_config,
  config -> 'replicas' -> 0 -> 'host' as first_replica_host,
  jsonb_array_length(config -> 'replicas') as replica_count,
  exists(
    select
      1
    from
      jsonb_array_elements(config -> 'features') as f
    where
      f ->> 'name' = 'ssl'
      and (f ->> 'enabled')::boolean = true
  ) as has_ssl_enabled,
  (
    select
      sum((r ->> 'priority')::integer)
    from
      jsonb_array_elements(config -> 'replicas') as r
  ) as total_replica_priority
from
  services
where
  config -> 'database' ->> 'type' = 'postgresql'
  and jsonb_array_length(config -> 'replicas') > 0
order by
  jsonb_array_length(config -> 'replicas') desc;

-- JSONB indexing and search operations
select
  document_id,
  title,
  content,
  content -> 'metadata' ->> 'author' as author,
  content -> 'metadata' ->> 'published_date' as published_date,
  content -> 'body' ->> 'summary' as summary,
  content -> 'tags' as tags,
  content @@ '$.metadata.author == "John Doe"' as authored_by_john,
  content @@ '$.tags[*] == "technology"' as has_technology_tag,
  content @@ '$.body.word_count > 1000' as is_long_article,
  content @? '$.comments[*].author ? (@ == "Jane Smith")' as commented_by_jane,
  jsonb_path_query(content, '$.comments[*].rating') as comment_ratings,
  jsonb_path_query_array(content, '$.tags[*]') as all_tags,
  jsonb_path_exists(content, '$.metadata.published_date') as has_publish_date,
  jsonb_path_match(content, '$.metadata.author == $author', '{"author": "John Doe"}') as matches_author
from
  documents
where
  content @@ '$.metadata.category == "technical"'
  and content @@ '$.body.word_count > 500'
order by
  content -> 'metadata' ->> 'published_date' desc;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/postgresql/advanced_jsonb_operations.sql