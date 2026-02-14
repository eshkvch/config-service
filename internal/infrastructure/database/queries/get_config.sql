SELECT env, key, value, updated_at
FROM configs
WHERE env = $1 AND key = $2;

