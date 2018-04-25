CREATE EXTENSION IF NOT EXISTS CITEXT;

drop table if exists users cascade;
drop table if exists forums cascade;
drop table if exists threads cascade;
drop table if exists messages cascade;
drop table if exists votes cascade;


CREATE TABLE IF NOT EXISTS users
(
  id       BIGSERIAL PRIMARY KEY,

  nickname VARCHAR(64) NOT NULL UNIQUE,
  email    CITEXT NOT NULL UNIQUE,

  about    TEXT DEFAULT '',
  fullname VARCHAR(96) DEFAULT ''
);


CREATE TABLE IF NOT EXISTS forums
(
  id      BIGSERIAL primary key,

  slug    CITEXT not null unique,

  title   CITEXT,

  threads INTEGER DEFAULT 0,
  posts   INTEGER DEFAULT 0,

  author  VARCHAR references users(nickname)
);

CREATE TABLE threads
(
  id         BIGSERIAL PRIMARY KEY,
  slug       CITEXT  not null UNIQUE,

  created    TIMESTAMP WITH TIME ZONE,

  message    TEXT,
  title      TEXT,

  author     VARCHAR REFERENCES users (nickname),
  forum      CITEXT REFERENCES forums(slug),

  votes      BIGINT DEFAULT 0
);

