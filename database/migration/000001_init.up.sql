CREATE TABLE IF NOT EXISTS "movies" (
  "id" text PRIMARY KEY,
  "title" text UNIQUE NOT NULL,
  "description" text NOT NULL,
  "director" text NOT NULL
);

CREATE TABLE IF NOT EXISTS "ratings" (
  "id" bigserial PRIMARY KEY,
  "movie_id" text NOT NULL,
  "record_type" text NOT NULL,
  "user_id" text NOT NULL,
  "value" integer NOT NULL
);

CREATE INDEX ON "ratings" ("movie_id", "record_type");