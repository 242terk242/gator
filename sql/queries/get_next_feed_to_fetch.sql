-- Emulating NULLS FIRST in MySQL/SQL Server (ASC order)
SELECT * FROM feeds
ORDER BY (CASE WHEN column_name IS NULL THEN 0 ELSE 1 END), column_name ASC;

-- Emulating NULLS FIRST in MySQL/SQL Server (DESC order)
--SELECT * FROM table_name
--ORDER BY (CASE WHEN column_name IS NULL THEN 0 ELSE 1 END), column_name DESC;