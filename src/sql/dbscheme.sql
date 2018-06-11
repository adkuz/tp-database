CREATE EXTENSION IF NOT EXISTS CITEXT;


drop table if exists users cascade;
drop table if exists forums cascade;
drop table if exists threads cascade;
drop table if exists posts cascade;
drop table if exists votes cascade;


drop index if exists threads_slug_idx;
drop index if exists forums_slug_idx;
drop index if exists users_slug_idx;


DROP INDEX IF EXISTS post_author_idx;
DROP INDEX IF EXISTS post_forum_idx;
DROP INDEX IF EXISTS post_thread_idx;
DROP INDEX IF EXISTS post_created_idx;
DROP INDEX IF EXISTS post_tree_parent_idx;
DROP INDEX IF EXISTS post_thread_created_id_idx;
DROP INDEX IF EXISTS post_created_thread_id_idx;
DROP INDEX IF EXISTS post_id_thread_idx;
DROP INDEX IF EXISTS post_thread_tree_path;


DROP INDEX IF EXISTS forum_users_username_idx;
DROP INDEX IF EXISTS forum_users_forum_slug_idx;


CREATE TABLE IF NOT EXISTS users
(
  nickname VARCHAR(64) NOT NULL UNIQUE primary key,
  email    CITEXT NOT NULL UNIQUE,

  about    TEXT DEFAULT '',
  fullname VARCHAR(96) DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS users_slug_idx ON users(lower(nickname));


CREATE TABLE IF NOT EXISTS forums
(
  id      BIGSERIAL primary key,

  slug    CITEXT not null unique,

  title   CITEXT,

  threads INTEGER DEFAULT 0,
  posts   INTEGER DEFAULT 0,

  author  VARCHAR references users(nickname)
);

CREATE UNIQUE INDEX IF NOT EXISTS forums_slug_idx ON forums(lower(slug));


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

CREATE UNIQUE INDEX IF NOT EXISTS threads_slug_idx ON threads(lower(slug));


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

CREATE INDEX IF NOT EXISTS post_author_idx ON posts(lower(author));
CREATE INDEX IF NOT EXISTS post_forum_idx ON posts(forum);
CREATE INDEX IF NOT EXISTS post_thread_idx ON posts(thread);
CREATE INDEX IF NOT EXISTS post_created_idx ON posts(created);
CREATE INDEX IF NOT EXISTS post_tree_parent_idx ON posts((tree_path[1]));
CREATE INDEX IF NOT EXISTS post_thread_created_id_idx ON posts(thread, created, id);
CREATE INDEX IF NOT EXISTS post_created_thread_id_idx ON posts(parent, thread, id);
CREATE INDEX IF NOT EXISTS post_id_thread_idx ON posts(id, thread);
CREATE INDEX IF NOT EXISTS post_thread_tree_path ON posts(thread, tree_path);

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
  forum CITEXT REFERENCES forums(slug) NOT NULL
);

CREATE INDEX IF NOT EXISTS forum_users_username_idx ON forum_users(lower(username));
CREATE INDEX IF NOT EXISTS forum_users_forum_slug_idx ON forum_users(forum);




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

