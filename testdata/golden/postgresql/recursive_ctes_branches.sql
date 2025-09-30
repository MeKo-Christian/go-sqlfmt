-- Recursive CTEs with multiple branches and complex recursion patterns
with recursive
  category_tree as(
    select
      id,
      name,
      parent_id,
      0 as level,
      array [id] as path,
      name as path_names
    from
      categories
    where
      parent_id is null
    union all
    select
      c.id,
      c.name,
      c.parent_id,
      ct.level + 1,
      ct.path || c.id,
      ct.path_names || ' > ' || c.name
    from
      categories c
      join category_tree ct on c.parent_id = ct.id
  ),
  product_hierarchy as(
    select
      p.id as product_id,
      p.name as product_name,
      p.category_id,
      ct.level as category_level,
      ct.path as category_path,
      ct.path_names as category_path_names
    from
      products p
      join category_tree ct on p.category_id = ct.id
  )
select
  ph.product_id,
  ph.product_name,
  ph.category_level,
  ph.category_path_names,
  array_length(ph.category_path, 1) as category_depth
from
  product_hierarchy ph
order by
  ph.category_path,
  ph.product_name;

-- Multiple recursive CTEs with interdependencies
with recursive
  employee_hierarchy as(
    select
      id,
      name,
      manager_id,
      department_id,
      salary,
      0 as level,
      array [id] as path
    from
      employees
    where
      manager_id is null
    union all
    select
      e.id,
      e.name,
      e.manager_id,
      e.department_id,
      e.salary,
      eh.level + 1,
      eh.path || e.id
    from
      employees e
      join employee_hierarchy eh on e.manager_id = eh.id
  ),
  department_budget as(
    select
      d.id as dept_id,
      d.name as dept_name,
      d.budget,
      count(e.id) as employee_count,
      sum(e.salary) as total_salary,
      d.budget - sum(e.salary) as remaining_budget
    from
      departments d
      left join employees e on d.id = e.department_id
      and e.active = true
    group by
      d.id,
      d.name,
      d.budget
  ),
  budget_analysis as(
    select
      eh.id as employee_id,
      eh.name as employee_name,
      eh.level as hierarchy_level,
      eh.salary,
      db.dept_name,
      db.total_salary as dept_total_salary,
      db.remaining_budget as dept_remaining_budget,
      eh.salary / nullif(db.total_salary, 0) * 100 as salary_percentage_of_dept,
      case
        when db.remaining_budget < 0 then 'over_budget'
        when db.remaining_budget < db.budget * 0.1 then 'near_budget_limit'
        else 'within_budget'
      end as budget_status
    from
      employee_hierarchy eh
      join department_budget db on eh.department_id = db.dept_id
  )
select
  ba.employee_name,
  ba.hierarchy_level,
  ba.salary,
  ba.dept_name,
  round(ba.salary_percentage_of_dept, 2) as salary_pct_of_dept,
  ba.budget_status,
  array_length(eh.path, 1) - 1 as reports_to_count
from
  budget_analysis ba
  join employee_hierarchy eh on ba.employee_id = eh.id
order by
  ba.dept_name,
  ba.hierarchy_level,
  ba.salary desc;

-- Recursive CTE with multiple branches and aggregation
with recursive
  bill_of_materials as(
    select
      component_id,
      component_name,
      parent_component_id,
      quantity,
      unit_cost,
      0 as level,
      array [component_id] as path,
      quantity * unit_cost as total_cost
    from
      components
    where
      parent_component_id is null
    union all
    select
      c.component_id,
      c.component_name,
      c.parent_component_id,
      c.quantity,
      c.unit_cost,
      bom.level + 1,
      bom.path || c.component_id,
      c.quantity * c.unit_cost
    from
      components c
      join bill_of_materials bom on c.parent_component_id = bom.component_id
  ),
  cost_rollup as(
    select
      parent_component_id,
      sum(total_cost) as total_subcomponent_cost,
      count(*) as subcomponent_count,
      avg(unit_cost) as avg_subcomponent_cost
    from
      bill_of_materials
    where
      level > 0
    group by
      parent_component_id
  )
select
  bom.component_name,
  bom.level,
  bom.quantity,
  bom.unit_cost,
  bom.total_cost,
  coalesce(cr.total_subcomponent_cost, 0) as subcomponent_cost,
  bom.total_cost + coalesce(cr.total_subcomponent_cost, 0) as total_cost_with_subs,
  coalesce(cr.subcomponent_count, 0) as subcomponent_count,
  array_length(bom.path, 1) as path_length
from
  bill_of_materials bom
  left join cost_rollup cr on bom.component_id = cr.parent_component_id
order by
  bom.path;

-- Complex recursive CTE with filtering and multiple recursion paths
with recursive
  flight_routes as(
    select
      f.flight_id,
      f.origin,
      f.destination,
      f.departure_time,
      f.arrival_time,
      f.price,
      0 as stops,
      array [f.origin, f.destination] as route,
      f.arrival_time - f.departure_time as duration,
      f.price as total_price,
      1 as segment_count
    from
      flights f
    where
      f.origin = 'JFK'
      and f.departure_time >= current_date
    union all
    select
      f.flight_id,
      fr.origin,
      f.destination,
      fr.departure_time,
      f.arrival_time,
      f.price,
      fr.stops + 1,
      fr.route || f.destination,
      f.arrival_time - fr.departure_time,
      fr.total_price + f.price,
      fr.segment_count + 1
    from
      flights f
      join flight_routes fr on f.origin = fr.destination
      and f.departure_time > fr.arrival_time
      and f.departure_time <= fr.arrival_time + interval '24 hours'
    where
      fr.stops < 3
      and f.destination != all(fr.route [1:array_length(fr.route, 1) - 1])
  )
select
  flight_id,
  origin,
  destination,
  route,
  stops,
  departure_time,
  arrival_time,
  duration,
  total_price,
  segment_count,
  total_price / segment_count as avg_price_per_segment
from
  flight_routes
where
  destination = 'LAX'
  and stops <= 2
order by
  total_price,
  duration;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/postgresql/recursive_ctes_branches.sql