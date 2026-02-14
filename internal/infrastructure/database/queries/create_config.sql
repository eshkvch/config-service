INSERT INTO configs (env, key, value, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (env, key)
DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;

