CREATE TABLE users (id SERIAL, user_id UUID NOT NULL, login TEXT NOT NULL UNIQUE, password TEXT NOT NULL);