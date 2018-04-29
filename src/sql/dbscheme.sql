CREATE EXTENSION IF NOT EXISTS CITEXT;

drop table if exists users cascade;
drop table if exists forums cascade;
drop table if exists threads cascade;
drop table if exists posts cascade;
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
  slug       CITEXT unique,

  created    TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,

  message    TEXT default '',
  title      TEXT default '',

  author     VARCHAR REFERENCES users (nickname),
  forum      CITEXT REFERENCES forums(slug),

  votes      INTEGER DEFAULT 0
);

create table if not exists posts
(
  id        bigserial not null primary key,

  created   TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,

  is_edited boolean default FALSE,

  parent    bigint DEFAULT 0,
  tree_path bigint array,

  message   text not null,

  author    varchar not null references users(nickname),
  forum     CITEXT references forums(slug),
  thread    bigint references threads(id)
);


CREATE TABLE votes
(
  id        bigserial   NOT NULL PRIMARY KEY,
  username  VARCHAR     NOT NULL REFERENCES users(nickname),
  thread    INTEGER     NOT NULL REFERENCES threads(id),
  voice     INTEGER,

  UNIQUE(username, thread)
);

CREATE TABLE forum_users
(
  username  VARCHAR REFERENCES users(nickname) NOT NULL,
  forum CITEXT REFERENCES forums(slug) NOT NULL,

  UNIQUE(username, forum)
);


CREATE FUNCTION fix_path() RETURNS trigger AS $fix_path$
DECLARE
  parent_id BIGINT;

BEGIN
  parent_id := new.parent;
  new.tree_path := array_append((SELECT tree_path from posts WHERE id = parent_id), new.id);
  UPDATE forums SET posts = posts + 1 WHERE LOWER(slug) = LOWER(new.forum);
  RETURN new;
END;
$fix_path$ LANGUAGE plpgsql;


CREATE TRIGGER fix_path BEFORE INSERT OR UPDATE ON posts
  FOR EACH ROW EXECUTE PROCEDURE fix_path();

