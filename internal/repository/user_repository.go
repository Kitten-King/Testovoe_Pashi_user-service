package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Kitten-King/user-sdk"
)

type UserRepository struct {
	db *sql.DB
}

type UserRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*user_sdk.UserWithCity, error)
	CreateUser(ctx context.Context, user *user_sdk.User) error
	FindWithinRadius(ctx context.Context, lat, lon, radius float64) ([]user_sdk.UserWithCity, error)
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *user_sdk.User) error {
	query := "INSERT INTO testovoe.user (name, city_id) VALUES ($1, $2) RETURNING user_id"
	return r.db.QueryRowContext(ctx, query, user.Name, user.CityID).Scan(&user.UserID)
}

func (r *UserRepository) UpdateUserCity(ctx context.Context, userID, cityID int) error {
	query := `UPDATE testovoe.user SET city_id = $1 WHERE user_id = $2`
	_, err := r.db.ExecContext(ctx, query, cityID, userID)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*user_sdk.UserWithCity, error) {
	var user user_sdk.UserWithCity
	query := "SELECT u.user_id, u.name, u.city_id, c.city_name, c.latitude, c.longitude FROM testovoe.user u JOIN testovoe.city c ON u.city_id = c.city_id WHERE u.user_id = $1"
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.UserID,
		&user.Name,
		&user.CityID,
		&user.CityName,
		&user.Latitude,
		&user.Longitude,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindWithinRadius(ctx context.Context, lat, lon, radiusKm float64) ([]user_sdk.UserWithCity, error) {
	var users []user_sdk.UserWithCity
	query := "SELECT u.user_id, u.name, u.city_id, c.city_name, c.latitude, c.longitude FROM testovoe.user u JOIN testovoe.city c ON u.city_id = c.city_id WHERE 6371 * acos( cos(radians($1)) * cos(radians(c.latitude)) * cos(radians(c.longitude) - radians($2)) + sin(radians($1)) * sin(radians(c.latitude))) <= $3"
	rows, err := r.db.QueryContext(ctx, query, lat, lon, radiusKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u user_sdk.UserWithCity
		if err := rows.Scan(&u.UserID, &u.Name, &u.CityID, &u.CityName, &u.Latitude, &u.Longitude); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
