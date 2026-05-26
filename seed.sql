TRUNCATE TABLE testovoe.user RESTART IDENTITY CASCADE;
TRUNCATE TABLE testovoe.city RESTART IDENTITY CASCADE;

WITH cities_source AS (
    SELECT * FROM (VALUES
                       ('Москва', 55.751244, 37.618423),
                       ('Санкт-Петербург', 59.934280, 30.335099),
                       ('Нью-Йорк', 40.712776, -74.005974),
                       ('Лондон', 51.507351, -0.127758),
                       ('Париж', 48.856613, 2.352222),
                       ('Токио', 35.689487, 139.691706),
                       ('Пекин', 39.904202, 116.407394),
                       ('Рим', 41.902782, 12.496366),
                       ('Берлин', 52.520007, 13.404954),
                       ('Мадрид', 40.416775, -3.703790),
                       ('Сидней', -33.868820, 151.209296),
                       ('Рио-де-Жанейро', -22.906847, -43.172896),
                       ('Кейптаун', -33.924868, 18.424055),
                       ('Дели', 28.613939, 77.209021),
                       ('Стамбул', 41.008238, 28.978359)
                  ) AS t(name, lat, lon)
)
INSERT INTO testovoe.city (city_name, latitude, longitude)
SELECT name, lat, lon
FROM (
         SELECT name, lat, lon
         FROM cities_source, generate_series(1, 50) AS idx
         ORDER BY random()
         LIMIT 50
     ) AS t;

INSERT INTO testovoe.user (name, city_id)
WITH available_cities AS (
    SELECT array_agg(city_id) as arr_ids FROM testovoe.city
)
SELECT
    'User_' || idx,
    arr_ids[floor(random() * array_length(arr_ids, 1) + 1)]
FROM generate_series(1, 500) AS idx
         CROSS JOIN available_cities;