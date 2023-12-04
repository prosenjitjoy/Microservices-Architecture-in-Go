-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2023-12-04T01:05:57.003Z

CREATE TABLE "movies" (
  "id" text PRIMARY KEY,
  "title" text UNIQUE NOT NULL,
  "description" text NOT NULL,
  "director" text NOT NULL
);

CREATE TABLE "ratings" (
  "id" bigserial PRIMARY KEY,
  "movie_id" text NOT NULL,
  "record_type" text NOT NULL,
  "user_id" text NOT NULL,
  "value" integer NOT NULL
);

CREATE INDEX ON "ratings" ("movie_id", "record_type");
