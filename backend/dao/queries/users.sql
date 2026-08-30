-- sqlc query definitions (repository layer uses pgx directly; queries reserved for codegen)

-- name: GetUserByEmail :one
SELECT id, email, hashed_password, role, is_active, display_name, created_at, updated_at
FROM users WHERE email = $1;
