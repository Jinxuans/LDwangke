@echo off
cd /d "%~dp0"
"C:\Program Files\MySQL\MySQL Server 9.4\bin\mysql.exe" -h127.0.0.1 -P3306 -u29_colnt_com -pifMezaaH5FEP31Z8 29_colnt_com < migrations/042_yfdk_projects_table.sql
echo Table created successfully!
pause
