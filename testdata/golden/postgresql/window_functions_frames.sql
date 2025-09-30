-- Complex window functions with various frame specifications
select
  employee_id,
  department,
  salary,
  hire_date,
  row_number() over(
    partition by department
    order by
      salary desc
  ) as dept_salary_rank,
  rank() over(
    partition by department
    order by
      salary desc
  ) as dept_salary_rank_dense,
  dense_rank() over(
    partition by department
    order by
      salary desc
  ) as dept_salary_rank_no_gaps,
  percent_rank() over(
    partition by department
    order by
      salary desc
  ) as dept_salary_percentile,
  cume_dist() over(
    partition by department
    order by
      salary desc
  ) as dept_salary_cumulative_dist,
  ntile(4) over(
    partition by department
    order by
      salary desc
  ) as dept_salary_quartile,
  lag(salary, 1, 0) over(
    partition by department
    order by
      hire_date
  ) as prev_salary_by_hire,
  lead(salary, 1, 0) over(
    partition by department
    order by
      hire_date
  ) as next_salary_by_hire,
  first_value(salary) over(
    partition by department
    order by
      hire_date rows between unbounded preceding
      and unbounded following
  ) as first_hired_salary,
  last_value(salary) over(
    partition by department
    order by
      hire_date rows between unbounded preceding
      and unbounded following
  ) as last_hired_salary,
  nth_value(salary, 2) over(
    partition by department
    order by
      hire_date rows between unbounded preceding
      and unbounded following
  ) as second_hired_salary,
  sum(salary) over(
    partition by department
    order by
      hire_date rows between unbounded preceding
      and current row
  ) as running_total_salary,
  avg(salary) over(
    partition by department
    order by
      hire_date rows between 2 preceding
      and 2 following
  ) as moving_avg_salary_5period,
  count(*) over(
    partition by department
    order by
      hire_date range between interval '1 year' preceding
      and interval '1 year' following
  ) as peers_hired_within_2years,
  min(salary) over(
    partition by department
    order by
      hire_date rows between unbounded preceding
      and current row
  ) as min_salary_to_date,
  max(salary) over(
    partition by department
    order by
      hire_date rows between current row
      and unbounded following
  ) as max_salary_future
from
  employees
where
  active = true
order by
  department,
  hire_date;

-- Advanced window functions with complex expressions and multiple windows
select
  product_id,
  product_name,
  category,
  sales_month,
  monthly_sales,
  sum(monthly_sales) over(
    partition by product_id
    order by
      sales_month rows between unbounded preceding
      and current row
  ) as cumulative_sales,
  avg(monthly_sales) over(
    partition by product_id
    order by
      sales_month rows between 3 preceding
      and current row
  ) as moving_avg_4months,
  sum(monthly_sales) over(
    partition by category,
    product_id
    order by
      sales_month range between interval '6 months' preceding
      and current row
  ) as category_product_sales_6months,
  row_number() over(
    partition by category,
    sales_month
    order by
      monthly_sales desc
  ) as monthly_category_rank,
  rank() over(
    partition by category
    order by
      sum(monthly_sales) over(
        partition by category,
        product_id
        order by
          sales_month rows between unbounded preceding
          and current row
      ) desc
  ) as overall_category_rank,
  lag(monthly_sales, 12) over(
    partition by product_id
    order by
      sales_month
  ) as sales_year_ago,
  (monthly_sales - lag(monthly_sales, 12, 0) over(
    partition by product_id
    order by
      sales_month
  )) / nullif(
    lag(monthly_sales, 12, 0) over(
      partition by product_id
      order by
      sales_month
    ),
    0
  ) * 100 as yoy_growth_percent
from
  product_sales
where
  sales_month >= '2020-01-01'
order by
  category,
  product_id,
  sales_month;

-- Window functions with GROUPS frame mode (PostgreSQL 10+)
select
  sensor_id,
  reading_time,
  temperature,
  avg(temperature) over(
    partition by sensor_id
    order by
      reading_time groups between 5 preceding
      and current row
  ) as avg_temp_6readings,
  sum(temperature) over(
    partition by sensor_id
    order by
      reading_time groups between unbounded preceding
      and 2 following
  ) as sum_temp_unbounded_to_2forward,
  count(*) over(
    partition by sensor_id
    order by
      reading_time groups between 10 preceding
      and 10 following
  ) as count_readings_21window,
  min(temperature) over(
    partition by sensor_id
    order by
      reading_time groups between current row
      and unbounded following
  ) as min_future_temp,
  max(temperature) over(
    partition by sensor_id
    order by
      reading_time groups between unbounded preceding
      and current row
  ) as max_past_temp
from
  sensor_readings
where
  reading_time >= current_timestamp - interval '24 hours'
order by
  sensor_id,
  reading_time;

-- Window functions with EXCLUDE clauses (PostgreSQL 11+)
select
  department,
  employee_id,
  salary,
  sum(salary) over(
    partition by department
    order by
      salary desc rows between unbounded preceding
      and unbounded following exclude current row
  ) as total_salary_excl_current,
  avg(salary) over(
    partition by department
    order by
      salary desc rows between unbounded preceding
      and unbounded following exclude group
  ) as avg_salary_excl_ties,
  count(*) over(
    partition by department
    order by
      salary desc rows between unbounded preceding
      and unbounded following exclude ties
  ) as count_excl_ties,
  rank() over(
    partition by department
    order by
      salary desc
  ) as salary_rank,
  dense_rank() over(
    partition by department
    order by
      salary desc
  ) as salary_dense_rank
from
  employees
where
  active = true
order by
  department,
  salary desc;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/postgresql/window_functions_frames.sql