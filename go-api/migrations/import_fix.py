import re

infile = r"c:\Users\29852\Desktop\1\149.88.74.83\7777_2026-02-06_13-08-22_mysql_data_wsplK.sql"
outfile = r"c:\Users\29852\Desktop\开发\29-colnt-com\go-api\migrations\cleaned_dump.sql"

with open(infile, "r", encoding="utf-8") as f:
    lines = f.readlines()

cleaned = []
skip = False
in_delimiter = False

for line in lines:
    stripped = line.strip()

    # Skip DEFINER lines
    if "DEFINER" in line:
        continue
    # Skip SET @@GLOBAL / SET @@SESSION
    if re.match(r"^\s*SET\s+@@", stripped, re.IGNORECASE):
        continue
    # Skip GRANT/REVOKE
    if re.match(r"^\s*(GRANT|REVOKE)\s+", stripped, re.IGNORECASE):
        continue

    # Track DELIMITER blocks (stored procedures, functions, triggers)
    if stripped.startswith("DELIMITER ;;") or stripped.startswith("DELIMITER  ;;"):
        in_delimiter = True
        continue
    if stripped.startswith("DELIMITER ;") and in_delimiter:
        in_delimiter = False
        continue
    if in_delimiter:
        continue

    # Skip conditional comments containing procedure/trigger setup
    if re.match(r"^/\*!50003\s+SET", stripped):
        continue
    if re.match(r"^/\*!50003\s+CREATE", stripped):
        continue

    cleaned.append(line)

with open(outfile, "w", encoding="utf-8") as f:
    f.writelines(cleaned)

print(f"Done. Original: {len(lines)} lines, Cleaned: {len(cleaned)} lines")
