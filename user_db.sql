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