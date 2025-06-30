CREATE TABLE devices (
     id SERIAL PRIMARY KEY,
     user_id integer NOT NULL REFERENCES users(id),
     device_name TEXT NOT NULL,
     last_sync TIMESTAMPTZ DEFAULT '2023-01-01 00:00:00+00'::TIMESTAMPTZ
);