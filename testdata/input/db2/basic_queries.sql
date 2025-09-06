select empno,lastname,firstname,salary,workdept from employee where workdept='A00'and salary>50000 order by salary desc,lastname asc;

select*from table(values('John','Doe',25,'Engineer'),('Jane','Smith',30,'Manager'),('Bob','Johnson',35,'Analyst'))as employees(firstname,lastname,age,position)where age>=30 order by age desc;