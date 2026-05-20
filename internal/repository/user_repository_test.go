package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"

	"github.com/Kitten-King/user-sdk"
)

func setupTestPostgres(t *testing.T) *sql.DB {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf("host=%s port=%s user=test password=test dbname=testdb sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	createTablesSQL := `
    CREATE SCHEMA IF NOT EXISTS testovoe;

    CREATE TABLE testovoe.city (
        city_id SERIAL PRIMARY KEY,
        city_name VARCHAR(100) NOT NULL,
        latitude FLOAT NOT NULL,
        longitude FLOAT NOT NULL
    );

    CREATE TABLE testovoe.user (
        user_id SERIAL PRIMARY KEY,
        name    VARCHAR(100) NOT NULL,
        city_id INTEGER      NOT NULL REFERENCES testovoe.city (city_id)
    );

    CREATE TABLE testovoe.trip (
        trip_id TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES testovoe.user(user_id) ON DELETE CASCADE,
        destination_city_id INTEGER NOT NULL REFERENCES testovoe.city(city_id) ON DELETE CASCADE,
        start_time TIMESTAMPTZ NOT NULL,
        duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
        status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
        cancellation_token TEXT NOT NULL UNIQUE
    );
    `
	_, err = db.Exec(createTablesSQL)
	require.NoError(t, err)

	return db
}

func TestUserRepository_Integration(t *testing.T) {
	db := setupTestPostgres(t)

	repo := NewUserRepository(db)

	ctx := context.Background()

	t.Run("create user and get with city", func(t *testing.T) {
		var cityID int
		err := db.QueryRow(`INSERT INTO testovoe.city (city_name, latitude, longitude) VALUES ($1, $2, $3) RETURNING city_id`,
			"TestCity", 55.75, 37.62).Scan(&cityID)
		require.NoError(t, err)

		newUser := &user_sdk.User{
			Name:   "IntegrationUser",
			CityID: cityID,
		}
		err = repo.CreateUser(ctx, newUser)
		require.NoError(t, err)
		require.NotZero(t, newUser.UserID, "CreateUser should set the UserID")

		fetchedUser, err := repo.GetByID(ctx, newUser.UserID)

		require.NoError(t, err)
		assert.Equal(t, "IntegrationUser", fetchedUser.Name)
		assert.Equal(t, "TestCity", fetchedUser.CityName)
	})
}
