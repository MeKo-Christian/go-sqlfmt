select employee_id,last_name,first_name,hire_date,salary,department_id from employees where department_id=10 and hire_date>=date'2020-01-01'order by hire_date desc;

select employee_id,last_name,level as hierarchy_level,sys_connect_by_path(last_name,'/')as hierarchy_path from employees start with manager_id is null connect by prior employee_id=manager_id order siblings by last_name;