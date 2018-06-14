CREATE EXTENSION IF NOT EXISTS CITEXT;


drop table if exists users cascade;
drop table if exists forums cascade;
drop table if exists threads cascade;
drop table if exists posts cascade;
drop table if exists votes cascade;
drop table if exists forum_users cascade;

drop index if exists threads_slug_idx;
drop index if exists threads_author_idx;
drop index if exists treads_forum_idx;
drop index if exists treads_forum_created_idx;

drop index if exists forums_slug_idx;
drop index if exists forums_author_idx;

drop index if exists users_slug_idx;
drop index if exists users_nickname_idx;


DROP INDEX IF EXISTS post_author_idx;
DROP INDEX IF EXISTS post_forum_idx;
DROP INDEX IF EXISTS post_thread_idx;
DROP INDEX IF EXISTS post_created_idx;
DROP INDEX IF EXISTS post_tree_parent_idx;
DROP INDEX IF EXISTS post_thread_created_id_idx;
DROP INDEX IF EXISTS post_created_thread_id_idx;
DROP INDEX IF EXISTS post_id_thread_idx;
DROP INDEX IF EXISTS post_thread_tree_path;

drop INDEX IF EXISTS votes_thread_username_idx;


DROP INDEX IF EXISTS forum_users_forum_username_idx;
DROP INDEX IF EXISTS forum_users_username_idx;
DROP INDEX IF EXISTS forum_users_forum_idx;



CREATE TABLE IF NOT EXISTS users
(
  nickname VARCHAR(64) NOT NULL UNIQUE primary key,
  email    CITEXT NOT NULL UNIQUE,

  about    TEXT DEFAULT '',
  fullname VARCHAR(96) DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users(lower(email));
CREATE UNIQUE INDEX IF NOT EXISTS users_nickname_idx ON users(lower(nickname));


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
CREATE INDEX IF NOT EXISTS forums_author_idx ON forums(lower(author));


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

CREATE INDEX IF NOT EXISTS treads_forum_idx ON threads(lower(forum));
CREATE INDEX IF NOT EXISTS treads_forum_created_idx ON threads(lower(forum), created);
CREATE INDEX IF NOT EXISTS threads_author_idx ON threads(lower(author));


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


CREATE INDEX IF NOT EXISTS post_tree_parent_idx ON posts((tree_path[1]));
CREATE INDEX IF NOT EXISTS post_thread_path_id_idx ON posts(thread, tree_path, id);
CREATE INDEX IF NOT EXISTS post_created_thread_id_idx ON posts(created, thread, id);

CREATE INDEX IF NOT EXISTS post_thread_id_idx ON posts(thread, id);
CREATE INDEX IF NOT EXISTS post_thread_created_id_idx ON posts(thread, created, id); --GetPostsFlat


CREATE TABLE votes
(
  id        bigserial   NOT NULL PRIMARY KEY,
  username  VARCHAR     NOT NULL REFERENCES users(nickname),
  thread    INTEGER     NOT NULL REFERENCES threads(id),
  voice     INTEGER,

  UNIQUE(username, thread)
);

CREATE UNIQUE INDEX IF NOT EXISTS votes_thread_username_idx ON votes(thread, lower(username));


CREATE TABLE forum_users
(
  username  VARCHAR REFERENCES users(nickname) NOT NULL,
  forum CITEXT REFERENCES forums(slug) NOT NULL,

  UNIQUE(forum, username)
);

CREATE UNIQUE INDEX IF NOT EXISTS forum_users_forum_username_idx ON forum_users(lower(forum), lower(username));
CREATE INDEX IF NOT EXISTS forum_users_username_idx ON forum_users(lower(username));
CREATE INDEX IF NOT EXISTS forum_users_forum_idx ON forum_users(lower(forum));



CREATE FUNCTION fix_path() RETURNS trigger AS $fix_path$
DECLARE
  parent_id BIGINT;

BEGIN
  parent_id := new.parent;
  new.tree_path := array_append((SELECT tree_path from posts WHERE id = parent_id), new.id);
 -- insert into forum_users (forum, username) values (new.forum, new.author) ON conflict (forum, username) do nothing;
  RETURN new;
END;
$fix_path$ LANGUAGE plpgsql;


CREATE TRIGGER fix_path BEFORE INSERT OR UPDATE ON posts
  FOR EACH ROW EXECUTE PROCEDURE fix_path();

