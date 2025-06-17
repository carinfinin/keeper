CREATE TABLE devices (
     id SERIAL PRIMARY KEY,
     user_id integer NOT NULL REFERENCES users(id),
     device_name TEXT NOT NULL,
     last_sync TIMESTAMPTZ
);