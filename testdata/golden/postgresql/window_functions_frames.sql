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
      hire_date rows between 3 preceding
      and current row
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