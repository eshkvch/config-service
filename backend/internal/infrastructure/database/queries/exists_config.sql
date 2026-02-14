SELECT EXISTS(
    SELECT 1
    FROM configs
    WHERE env = $1 AND key = $2
);

