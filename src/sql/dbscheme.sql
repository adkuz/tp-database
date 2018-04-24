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

  author_id  BIGINT references users(id),

  title   CITEXT,

  threads INTEGER DEFAULT 0,
  posts   INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS threads
(
  id         BIGSERIAL PRIMARY KEY,
  slug       TEXT UNIQUE,
  created_on TIMESTAMP,
  message    TEXT,
  title      TEXT,
  authorid   BIGINT REFERENCES users (id),
  forumid    BIGINT REFERENCES forums (id)
);

CREATE TABLE IF NOT EXISTS messages
(
  id         BIGSERIAL PRIMARY KEY,
  created_on TIMESTAMP,
  message    TEXT,
  isedited   BOOLEAN,
  authorid   BIGINT REFERENCES users (id),
  parentid   BIGINT REFERENCES messages (id) DEFAULT 0,
  threadid   BIGINT REFERENCES threads (id),
  forumid    BIGINT REFERENCES forums (id),
  parenttree BIGINT[] DEFAULT '{0}'
);

CREATE TABLE IF NOT EXISTS votes
(
  voice      INT CHECK (voice in (1, -1)),
  userid     BIGINT REFERENCES users (id),
  threadid   BIGINT REFERENCES threads (id),

  CONSTRAINT unique_vote UNIQUE (userid, threadid)
);