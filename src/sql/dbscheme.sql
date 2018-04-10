CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS forums CASCADE;
DROP TABLE IF EXISTS threads CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS votes CASCADE;


CREATE TABLE IF NOT EXISTS users
(
  id       BIGSERIAL PRIMARY KEY,
  nickname CITEXT COLLATE ucs_basic NOT NULL UNIQUE,
  about    CITEXT DEFAULT '',
  email    CITEXT NOT NULL UNIQUE,
  fullname TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS forums
(
  id BIGSERIAL PRIMARY KEY,
  slug citext not null unique,
  title text default '',
  userid bigint references users(id)
);

CREATE TABLE IF NOT EXISTS threads
(
  id         BIGSERIAL PRIMARY KEY,
  slug       CITEXT UNIQUE,
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