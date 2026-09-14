-- Module 6 / SQL: nearest points (PostGIS) + Haversine fallback

-- PostGIS KNN
SELECT
    name,
    ST_Distance(location, ST_MakePoint(37.6156, 55.7522)::geography) / 1000 AS distance_km
FROM suppliers
ORDER BY location <-> ST_MakePoint(37.6156, 55.7522)
LIMIT 10;

-- Без PostGIS (Haversine, км)
SELECT
    name,
    (6371 * acos(
        cos(radians(55.7522)) * cos(radians(lat))
        * cos(radians(lon) - radians(37.6156))
        + sin(radians(55.7522)) * sin(radians(lat))
    )) AS distance_km
FROM suppliers
ORDER BY distance_km
LIMIT 10;

-- На собесе: индекс GiST/SP-GiST на geography; Ball Tree / S2 — app-side geo.
