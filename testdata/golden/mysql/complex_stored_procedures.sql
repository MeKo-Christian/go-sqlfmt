DELIMITER
/ /
CREATE PROCEDURE
  ProcessMonthlyReport(IN report_month DATE, IN report_year INT) BEGIN
    DECLARE current_dept_id INT;

DECLARE dept_done BOOLEAN DEFAULT FALSE;

DECLARE total_employees INT DEFAULT 0;

DECLARE total_salary DECIMAL(15, 2) DEFAULT 0;

DECLARE avg_performance DECIMAL(5, 2) DEFAULT 0;

DECLARE dept_cursor CURSOR FOR
SELECT
  id
FROM
  departments
WHERE
  active = 1
ORDER BY
  name;

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  dept_done = TRUE;

CREATE TEMPORARY TABLE monthly_report(
  department_id INT,
  department_name VARCHAR(100),
  employee_count INT,
  total_salary DECIMAL(15, 2),
  avg_salary DECIMAL(12, 2),
  top_performer VARCHAR(100),
  performance_score DECIMAL(5, 2),
  budget_utilization DECIMAL(5, 2)
);

OPEN dept_cursor;

dept_loop: LOOP
  FETCH dept_cursor INTO current_dept_id;

IF
  dept_done THEN
  LEAVE dept_loop;

END IF
;

SELECT
  COUNT(e.id),
  COALESCE(SUM(e.salary), 0),
  COALESCE(AVG(e.salary), 0),
  COALESCE(AVG(p.score), 0) INTO total_employees,
  total_salary,
  @avg_salary,
  avg_performance
FROM
  employees e
  LEFT JOIN performance_reviews p ON e.id = p.employee_id
  AND YEAR(p.review_date) = report_year
  AND MONTH(p.review_date) = MONTH(report_month)
WHERE
  e.department_id = current_dept_id
  AND e.active = 1;

SELECT
  COALESCE(e.name, 'N/A') INTO @top_performer
FROM
  employees e
  JOIN performance_reviews p ON e.id = p.employee_id
WHERE
  e.department_id = current_dept_id
  AND YEAR(p.review_date) = report_year
  AND MONTH(p.review_date) = MONTH(report_month)
ORDER BY
  p.score DESC
LIMIT
  1;

SELECT
  COALESCE((total_salary / NULLIF(d.budget, 0)) * 100, 0) INTO @budget_util
FROM
  departments d
WHERE
  d.id = current_dept_id;

INSERT INTO
  monthly_report(
    department_id,
    department_name,
    employee_count,
    total_salary,
    avg_salary,
    top_performer,
    performance_score,
    budget_utilization
  )
SELECT
  d.id,
  d.name,
  total_employees,
  total_salary,
  @avg_salary,
  @top_performer,
  avg_performance,
  @budget_util
FROM
  departments d
WHERE
  d.id = current_dept_id;

END LOOP
  dept_loop;

CLOSE dept_cursor;

SELECT
  *
FROM
  monthly_report
ORDER BY
  total_salary DESC;

DROP TEMPORARY TABLE monthly_report;

END / / DELIMITER
;