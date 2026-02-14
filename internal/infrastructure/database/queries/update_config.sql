UPDATE configs
SET value = $3, updated_at = $4
WHERE env = $1 AND key = $2;

