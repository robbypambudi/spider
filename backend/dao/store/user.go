package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID             uuid.UUID
	Email          string
	HashedPassword string
	Role           string
	IsActive       bool
	DisplayName    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, hashed_password, role, is_active, display_name, created_at, updated_at
		FROM users WHERE email = $1`, email)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.HashedPassword, &u.Role, &u.IsActive, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, hashed_password, role, is_active, display_name, created_at, updated_at
		FROM users WHERE id = $1`, id)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.HashedPassword, &u.Role, &u.IsActive, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, email, hashedPassword, role string, displayName *string) (*User, error) {
	now := time.Now().UTC()
	id := uuid.New()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, hashed_password, role, is_active, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, $5, $6, $7)`,
		id, email, hashedPassword, role, displayName, now, now)
	if err != nil {
		return nil, err
	}
	return &User{
		ID: id, Email: email, HashedPassword: hashedPassword, Role: role,
		IsActive: true, DisplayName: displayName, CreatedAt: now, UpdatedAt: now,
	}, nil
}
