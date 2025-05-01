-- Insert test data for go-sqlcl testing
-- This script assumes the tables have been created using create_tables.sql

-- Clear existing data to ensure idempotence
DELETE FROM job_history;
DELETE FROM employees;
DELETE FROM departments;

-- Insert departments
INSERT INTO departments (department_id, department_name, location) VALUES (10, 'Administration', 'Seattle');
INSERT INTO departments (department_id, department_name, location) VALUES (20, 'Marketing', 'New York');
INSERT INTO departments (department_id, department_name, location) VALUES (30, 'Purchasing', 'Chicago');
INSERT INTO departments (department_id, department_name, location) VALUES (40, 'Human Resources', 'Portland');
INSERT INTO departments (department_id, department_name, location) VALUES (50, 'IT', 'San Francisco');

-- Insert employees
-- Note: Insert managers first to satisfy foreign key constraints
INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (100, 'Steven', 'King', 'sking@example.com', '515.123.4567', TO_DATE('2003-06-17', 'YYYY-MM-DD'), 'President', 24000, NULL, NULL, 10);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (101, 'Neena', 'Kochhar', 'nkochhar@example.com', '515.123.4568', TO_DATE('2005-09-21', 'YYYY-MM-DD'), 'Vice President', 17000, NULL, 100, 10);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (102, 'Lex', 'De Haan', 'ldehaan@example.com', '515.123.4569', TO_DATE('2001-01-13', 'YYYY-MM-DD'), 'Vice President', 17000, NULL, 100, 10);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (201, 'Michael', 'Hartstein', 'mhartste@example.com', '515.123.5555', TO_DATE('2004-02-17', 'YYYY-MM-DD'), 'Marketing Manager', 13000, NULL, 100, 20);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (202, 'Pat', 'Fay', 'pfay@example.com', '603.123.6666', TO_DATE('2005-08-17', 'YYYY-MM-DD'), 'Marketing Representative', 6000, NULL, 201, 20);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (114, 'Raphaely', 'Den', 'drapheal@example.com', '515.127.4561', TO_DATE('2002-12-07', 'YYYY-MM-DD'), 'Purchasing Manager', 11000, NULL, 100, 30);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (115, 'Alexander', 'Khoo', 'akhoo@example.com', '515.127.4562', TO_DATE('2003-05-18', 'YYYY-MM-DD'), 'Purchasing Clerk', 3100, NULL, 114, 30);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (203, 'Susan', 'Mavris', 'smavris@example.com', '515.123.7777', TO_DATE('2002-06-07', 'YYYY-MM-DD'), 'HR Manager', 6500, NULL, 101, 40);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (103, 'Alexander', 'Hunold', 'ahunold@example.com', '590.423.4567', TO_DATE('2006-01-03', 'YYYY-MM-DD'), 'IT Director', 9000, NULL, 102, 50);

INSERT INTO employees (employee_id, first_name, last_name, email, phone_number, hire_date, job_title, salary, commission_pct, manager_id, department_id)
VALUES (104, 'Bruce', 'Ernst', 'bernst@example.com', '590.423.4568', TO_DATE('2007-05-21', 'YYYY-MM-DD'), 'Developer', 6000, NULL, 103, 50);

-- Insert job history
INSERT INTO job_history (employee_id, start_date, end_date, job_title, department_id)
VALUES (101, TO_DATE('2001-10-28', 'YYYY-MM-DD'), TO_DATE('2005-03-15', 'YYYY-MM-DD'), 'HR Assistant', 40);

INSERT INTO job_history (employee_id, start_date, end_date, job_title, department_id)
VALUES (101, TO_DATE('2005-03-16', 'YYYY-MM-DD'), TO_DATE('2007-10-19', 'YYYY-MM-DD'), 'HR Manager', 40);

INSERT INTO job_history (employee_id, start_date, end_date, job_title, department_id)
VALUES (102, TO_DATE('2001-01-13', 'YYYY-MM-DD'), TO_DATE('2006-07-24', 'YYYY-MM-DD'), 'IT Support', 50);

INSERT INTO job_history (employee_id, start_date, end_date, job_title, department_id)
VALUES (201, TO_DATE('2004-02-17', 'YYYY-MM-DD'), TO_DATE('2007-12-19', 'YYYY-MM-DD'), 'Marketing Representative', 20);

COMMIT; 