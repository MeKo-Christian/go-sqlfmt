-- Multi-CTE queries with RECURSIVE
with
  recursive employee_hierarchy as(
    select
      emp_id,
      name,
      manager_id,
      department,
      salary,
      0 as level
    from
      employees
    where
      manager_id is null
    union all
    select
      e.emp_id,
      e.name,
      e.manager_id,
      e.department,
      e.salary,
      eh.level + 1
    from
      employees e
      join employee_hierarchy eh on e.manager_id = eh.emp_id
  ),
  department_stats as(
    select
      department,
      count(*) as total_employees,
      avg(salary) as avg_salary,
      max(level) as max_hierarchy_depth
    from
      employee_hierarchy
    group by
      department
  )
select
  eh.name,
  eh.department,
  eh.level,
  eh.salary,
  ds.avg_salary,
  ds.max_hierarchy_depth
from
  employee_hierarchy eh
  join department_stats ds on eh.department = ds.department
where
  eh.salary > ds.avg_salary
order by
  eh.department,
  eh.level,
  eh.name;

-- UPSERT queries (INSERT...ON CONFLICT) - known formatting issues
insert into
  user_profiles(
    user_id,
    name,
    email,
    bio,
    preferences,
    created_at,
    updated_at
  )
values
(
    1,
    'John Doe',
    'john@example.com',
    'Software engineer with 10 years experience',
    '{"theme":"dark","notifications":true}',
    now(),
    now()
  ) on conflict(user_id)
do
update
set
  name = excluded.name,
  email = excluded.email,
  bio = excluded.bio,
  preferences = excluded.preferences,
  updated_at = now()
returning
  user_id,
  name,
  email,
  updated_at;

insert into
  product_inventory(
    sku,
    product_name,
    category,
    price,
    quantity,
    last_updated
  )
values
(
    'ABC123',
    'Premium Widget',
    'Electronics',
    299.99,
    50,
    now()
  ),
(
    'DEF456',
    'Standard Widget',
    'Electronics',
    199.99,
    100,
    now()
  ),
(
    'GHI789',
    'Basic Widget',
    'Electronics',
    99.99,
    200,
    now()
  ) on conflict(sku)
do
update
set
  price = excluded.price,
  quantity = inventory.quantity + excluded.quantity,
  last_updated = now()
where
  inventory.last_updated < excluded.last_updated
returning
  sku,
  product_name,
  price,
  quantity;

-- Complex JOIN queries with LATERAL
select
  u.user_id,
  u.username,
  u.email,
  recent_orders.order_count,
  recent_orders.total_spent,
  recent_orders.last_order_date,
  top_categories.category_name,
  top_categories.category_orders
from
  users u
  left join lateral(
    select
      count(*) as order_count,
      sum(total_amount) as total_spent,
      max(order_date) as last_order_date
    from
      orders o
    where
      o.user_id = u.user_id
      and o.order_date >= current_date - interval '90 days'
  ) recent_orders on true
  left join lateral(
    select
      c.category_name,
      count(oi.order_item_id) as category_orders
    from
      orders o
      join order_items oi on o.order_id = oi.order_id
      join products p on oi.product_id = p.product_id
      join categories c on p.category_id = c.category_id
    where
      o.user_id = u.user_id
      and o.order_date >= current_date - interval '365 days'
    group by
      c.category_id,
      c.category_name
    order by
      count(oi.order_item_id) desc
    limit
      1
  ) top_categories on true
where
  u.active = true
  and(
    recent_orders.order_count > 0
    or u.created_at >= current_date - interval '30 days'
  )
order by
  recent_orders.total_spent desc nulls last,
  u.username;

select
  stores.store_id,
  stores.store_name,
  stores.city,
  nearby_competitors.competitor_count,
  nearby_competitors.avg_distance,
  sales_analysis.monthly_revenue,
  sales_analysis.top_product
from
  stores
  cross join lateral(
    select
      count(*) as competitor_count,
      avg(st_distance(stores.location, comp.location)) as avg_distance
    from
      stores comp
    where
      comp.store_id != stores.store_id
      and st_dwithin(stores.location, comp.location, 5000)
  ) nearby_competitors
  cross join lateral(
    select
      sum(s.total_amount) as monthly_revenue,
      p.product_name as top_product
    from
      sales s
      join products p on s.product_id = p.product_id
    where
      s.store_id = stores.store_id
      and s.sale_date >= date_trunc('month', current_date)
    group by
      p.product_id,
      p.product_name
    order by
      sum(s.total_amount) desc
    limit
      1
  ) sales_analysis
where
  stores.active = true
order by
  sales_analysis.monthly_revenue desc;

-- Aggregate functions with FILTER clauses
select
  department,
  region,
  count(*) as total_employees,
  count(*) filter(
    where
      salary >= 50000
  ) as high_earners,
  count(*) filter(
    where
      hire_date >= current_date - interval '1 year'
  ) as recent_hires,
  avg(salary) filter(
    where
      performance_rating = 'excellent'
  ) as excellent_avg_salary,
  sum(salary) filter(
    where
      department = 'engineering'
  ) as engineering_total_cost,
  max(salary) filter(
    where
      gender = 'female'
  ) as highest_female_salary,
  percentile_cont(0.5) within group(
    order by
      salary
  ) filter(
    where
      active = true
  ) as median_active_salary
from
  employees
where
  active = true
group by
  department,
  region
having
  count(*) filter(
    where
      salary >= 50000
  ) > 5
order by
  count(*) filter(
    where
      salary >= 50000
  ) desc;

select
  product_category,
  extract(
    year
    from
      sale_date
  ) as sale_year,
  extract(
    month
    from
      sale_date
  ) as sale_month,
  sum(quantity * unit_price) as total_revenue,
  count(distinct customer_id) filter(
    where
      customer_type = 'premium'
  ) as premium_customers,
  avg(quantity * unit_price) filter(
    where
      discount_percent = 0
  ) as avg_full_price_sale,
  count(*) filter(
    where
      quantity > 10
  ) as bulk_orders,
  sum(quantity * unit_price) filter(
    where
      sale_date = date_trunc('month', sale_date) + interval '1 month' - interval '1 day'
  ) as month_end_revenue
from
  sales s
  join products p on s.product_id = p.product_id
where
  sale_date >= current_date - interval '2 years'
group by
  rollup(
    product_category,
    extract(
      year
      from
        sale_date
    ),
    extract(
      month
      from
        sale_date
    )
  )
having
  sum(quantity * unit_price) > 1000
order by
  sale_year desc nulls last,
  sale_month desc nulls last,
  total_revenue desc;

-- Window functions with complex frame specifications and JSON/JSONB operations
select
  customer_id,
  order_date,
  order_data ->> 'order_type' as order_type,
  order_data -> 'items' as order_items,
  jsonb_array_length(order_data -> 'items') as item_count,
(order_data ->> 'total_amount')::numeric as total_amount,
  row_number() over(
    partition by customer_id
    order by
      order_date desc
  ) as order_rank,
  lag((order_data ->> 'total_amount')::numeric, 1, 0) over(
    partition by customer_id
    order by
      order_date
  ) as prev_order_amount,
  sum((order_data ->> 'total_amount')::numeric) over(
    partition by customer_id
    order by
      order_date rows between unbounded preceding
      and current row
  ) as running_total,
  avg((order_data ->> 'total_amount')::numeric) over(
    partition by customer_id
    order by
      order_date rows between 2 preceding
      and current row
  ) as moving_avg_3orders,
  first_value(order_data ->> 'order_type') over(
    partition by customer_id
    order by
      order_date rows between unbounded preceding
      and unbounded following
  ) as first_order_type,
  count(*) filter(
    where
(order_data ->> 'total_amount')::numeric > 100
  ) over(
    partition by customer_id
    order by
      order_date range between interval '30 days' preceding
      and current row
  ) as recent_large_orders
from
  orders
where
  order_data?'total_amount'and order_data?'items'and jsonb_typeof(order_data -> 'items') = 'array'
  and(order_data ->> 'status')::text = 'completed'
order by
  customer_id,
(order_data ->> 'total_amount')::numeric desc;

-- Array operations, type casting, and pattern matching
select
  user_id,
  username,
  email,
  user_tags,
  array_length(user_tags, 1) as tag_count,
  user_preferences::jsonb ->> 'theme' as preferred_theme,
case
    when email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' then 'valid_email'
    else 'invalid_email'
  end as email_validity,
case
    when username ilike '%admin%'
    or username ilike '%root%' then true
    else false
  end as is_admin_user,
  array_to_string(user_tags[1:3 ], ', ') as first_three_tags,
  user_tags & & array ['premium',
  'vip' ]as has_special_status
from
  users
where
  array_length(user_tags, 1) > 0
  and user_preferences::text ~ 'theme'
  and email similar to '%@(gmail|yahoo|hotmail).com'
  and not(username ~* '^(test|demo|sample)')
order by
  array_length(user_tags, 1) desc,
  username;

-- Complex analytical query with numbered placeholders and dollar-quoted strings
select
  analytical_function($1::text, $2::integer) as result,
  execute_dynamic_query(
    $sql$select count(*)from information_schema.tables where table_schema not in('information_schema','pg_catalog')$sql$
  ) as user_table_count,
case
    when $3::boolean then $sql$Complex multi-line string with 'quotes' and "double quotes" that demonstrates dollar quoting with $nested$ content $nested$ inside$sql$
    else 'Simple string'
  end as conditional_text
from
(
    select
      row_number() over(
        order by
          random()
      ) as rn,
      generate_series(1, coalesce($4::integer, 100)) as series_value
  ) data
where
  rn between $5::integer
  and $6::integer
order by
  rn;

-- DO blocks and PL/pgSQL functions with complex nesting
do
  $complex_block$ declare user_count integer:=0;avg_salary numeric(10,2);dept_cursor cursor for select department_id,department_name from departments where active=true;dept_record record;begin for dept_record in dept_cursor loop select count(*),avg(salary) into user_count,avg_salary from employees where department_id=dept_record.department_id and active=true;if user_count>0 then raise notice'Department: % has % employees with average salary: $%',dept_record.department_name,user_count,avg_salary;insert into department_statistics(department_id,employee_count,average_salary,analysis_date)values(dept_record.department_id,user_count,avg_salary,current_timestamp)on conflict(department_id,analysis_date::date)do update set employee_count=excluded.employee_count,average_salary=excluded.average_salary;end if;end loop;end $complex_block$;

create or replace function
  calculate_employee_bonus(
    p_employee_id integer,
    p_performance_multiplier numeric default 1.0,
    p_department_bonus_pool numeric default null
  ) returns table(
    employee_id integer,
    base_salary numeric,
    performance_score numeric,
    calculated_bonus numeric,
    bonus_percentage numeric
  ) as $bonus_calculation$ declare emp_record record;dept_avg_salary numeric;company_performance numeric;begin select e.id,e.name,e.salary,e.department_id,p.score into emp_record from employees e left join performance_reviews p on e.id=p.employee_id where e.id=p_employee_id and e.active=true;if not found then raise exception'Employee % not found or inactive',p_employee_id;end if;select avg(salary) into dept_avg_salary from employees where department_id=emp_record.department_id and active=true;select performance_score into company_performance from company_metrics where metric_year=extract(year from current_date)and metric_type='overall_performance';return query with bonus_calculations as(select emp_record.id as emp_id,emp_record.salary,coalesce(emp_record.score,3.0)as perf_score,case when coalesce(emp_record.score,3.0)>=4.5 then emp_record.salary*0.15*p_performance_multiplier when coalesce(emp_record.score,3.0)>=3.5 then emp_record.salary*0.10*p_performance_multiplier when coalesce(emp_record.score,3.0)>=2.5 then emp_record.salary*0.05*p_performance_multiplier else 0 end+coalesce(p_department_bonus_pool/nullif((select count(*)from employees where department_id=emp_record.department_id and active=true),0),0)as bonus_amount)select bc.emp_id,bc.salary,bc.perf_score,bc.bonus_amount,round((bc.bonus_amount/bc.salary)*100,2)from bonus_calculations bc;end $bonus_calculation$ language plpgsql stable security definer cost 100;