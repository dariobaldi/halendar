CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    username citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL DEFAULT FALSE,
    access_level INTEGER NOT NULL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);
