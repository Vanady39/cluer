UPDATE apps
SET allowed_origins = ARRAY(
    SELECT DISTINCT rtrim(lower(o), '/')
    FROM unnest(allowed_origins) AS o
)
WHERE allowed_origins <> '{}';