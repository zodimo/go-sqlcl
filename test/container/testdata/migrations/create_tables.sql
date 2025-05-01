-- Create test tables for go-sqlcl testing
-- This script is designed to be idempotent

-- Drop existing objects first to ensure clean state
BEGIN
   -- Drop tables if they exist
   FOR t IN (SELECT table_name FROM user_tables WHERE table_name IN ('EMPLOYEES', 'DEPARTMENTS', 'JOB_HISTORY')) LOOP
      EXECUTE IMMEDIATE 'DROP TABLE ' || t.table_name || ' CASCADE CONSTRAINTS';
   END LOOP;
   
   -- Drop sequences if they exist
   FOR s IN (SELECT sequence_name FROM user_sequences WHERE sequence_name IN ('EMP_SEQ', 'DEPT_SEQ')) LOOP
      EXECUTE IMMEDIATE 'DROP SEQUENCE ' || s.sequence_name;
   END LOOP;
EXCEPTION
   WHEN OTHERS THEN
      NULL; -- Ignore errors during cleanup
END;
/

-- Create DEPARTMENTS table
CREATE TABLE departments (
    department_id   NUMBER(4) CONSTRAINT dept_pk PRIMARY KEY,
    department_name VARCHAR2(30) NOT NULL,
    location        VARCHAR2(50)
);

-- Create EMPLOYEES table
CREATE TABLE employees (
    employee_id    NUMBER(6) CONSTRAINT emp_pk PRIMARY KEY,
    first_name     VARCHAR2(20),
    last_name      VARCHAR2(25) NOT NULL,
    email          VARCHAR2(25) NOT NULL CONSTRAINT emp_email_uk UNIQUE,
    phone_number   VARCHAR2(20),
    hire_date      DATE NOT NULL,
    job_title      VARCHAR2(35) NOT NULL,
    salary         NUMBER(8,2),
    commission_pct NUMBER(2,2),
    manager_id     NUMBER(6) CONSTRAINT emp_mgr_fk REFERENCES employees,
    department_id  NUMBER(4) CONSTRAINT emp_dept_fk REFERENCES departments
);

-- Create JOB_HISTORY table
CREATE TABLE job_history (
    employee_id    NUMBER(6),
    start_date     DATE,
    end_date       DATE NOT NULL,
    job_title      VARCHAR2(35) NOT NULL,
    department_id  NUMBER(4),
    CONSTRAINT jhist_pk PRIMARY KEY (employee_id, start_date),
    CONSTRAINT jhist_emp_fk FOREIGN KEY (employee_id) REFERENCES employees,
    CONSTRAINT jhist_dept_fk FOREIGN KEY (department_id) REFERENCES departments,
    CONSTRAINT jhist_date_check CHECK (end_date > start_date)
);

-- Create sequences
CREATE SEQUENCE emp_seq START WITH 1 INCREMENT BY 1 NOCACHE;
CREATE SEQUENCE dept_seq START WITH 1 INCREMENT BY 1 NOCACHE;

-- Create trigger for employee_id
CREATE OR REPLACE TRIGGER emp_id_trigger
BEFORE INSERT ON employees
FOR EACH ROW
BEGIN
    IF :new.employee_id IS NULL THEN
        SELECT emp_seq.NEXTVAL INTO :new.employee_id FROM dual;
    END IF;
END;
/

-- Create trigger for department_id
CREATE OR REPLACE TRIGGER dept_id_trigger
BEFORE INSERT ON departments
FOR EACH ROW
BEGIN
    IF :new.department_id IS NULL THEN
        SELECT dept_seq.NEXTVAL INTO :new.department_id FROM dual;
    END IF;
END;
/ 