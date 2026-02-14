SELECT env, key, value, updated_at
FROM configs
WHERE env = $1
ORDER BY key;

