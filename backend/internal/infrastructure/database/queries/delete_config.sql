DELETE FROM configs
WHERE env = $1 AND key = $2;

