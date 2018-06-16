CREATE EXTENSION IF NOT EXISTS CITEXT;


drop table if exists users cascade;
drop table if exists forums cascade;
drop table if exists threads cascade;
drop table if exists posts cascade;
drop table if exists votes cascade;
drop table if exists forum_users cascade;



CREATE OR REPLACE FUNCTION drop_all_indexes() RETURNS INTEGER AS $$
DECLARE
  i RECORD;
BEGIN
  FOR i IN 
    (SELECT relname FROM pg_class
       -- exclude all pkey, exclude system catalog which starts with 'pg_'
      WHERE relkind = 'i' AND relname NOT LIKE '%_pkey%' AND relname NOT LIKE 'pg_%')
  LOOP
    -- RAISE INFO 'DROPING INDEX: %', i.relname;
    EXECUTE 'DROP INDEX ' || i.relname;
  END LOOP;
RETURN 1;
END;
$$ LANGUAGE plpgsql;

SELECT drop_all_indexes();


DROP INDEX IF EXISTS users_email_idx;
DROP INDEX IF EXISTS users_nickname_idx;
DROP INDEX IF EXISTS forums_slug_idx;
DROP INDEX IF EXISTS forums_author_idx;
DROP INDEX IF EXISTS threads_slug_idx;
DROP INDEX IF EXISTS treads_forum_idx;
DROP INDEX IF EXISTS treads_forum_created_idx;
DROP INDEX IF EXISTS treads_created_forum_idx;
DROP INDEX IF EXISTS threads_author_idx;
DROP INDEX IF EXISTS post_id_root_idx;
DROP INDEX IF EXISTS post_root_path_id_parent_id_idx;
DROP INDEX IF EXISTS post_path_id_parent_id_idx;
DROP INDEX IF EXISTS post_thread_parent_path_id_idx;
DROP INDEX IF EXISTS post_root_idx;
DROP INDEX IF EXISTS post_thread_parent_id_idx;
DROP INDEX IF EXISTS post_thread_path_id_idx;
DROP INDEX IF EXISTS post_path_id_idx;
DROP INDEX IF EXISTS post_thread_created_id_idx;
DROP INDEX IF EXISTS post_thread_id_idx;
DROP INDEX IF EXISTS votes_thread_username_idx;
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
CREATE INDEX IF NOT EXISTS treads_created_forum_idx ON threads(created, lower(forum));

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

-- root finding
CREATE INDEX IF NOT EXISTS post_id_root_idx ON posts(id, (tree_path[1]));

-- for parent_tree_sort desc?
CREATE INDEX IF NOT EXISTS post_root_path_id_parent_id_idx ON posts((tree_path[1]) DESC, tree_path, id);
-- for parent_tree_sort asc?
CREATE INDEX IF NOT EXISTS post_path_id_parent_id_idx ON posts(tree_path, id);

-- parent_tree_sort: parent selection
CREATE INDEX IF NOT EXISTS post_thread_parent_path_id_idx ON posts(thread, parent, tree_path);

-- root index in the thread
CREATE INDEX IF NOT EXISTS post_root_idx ON posts(thread, (tree_path[1]));

-- tree_sort: pre-selection
CREATE INDEX IF NOT EXISTS post_thread_parent_id_idx ON posts(id, tree_path); --checked
CREATE INDEX IF NOT EXISTS post_thread_path_id_idx ON posts(thread, tree_path);

-- tree_sort: sort
CREATE INDEX IF NOT EXISTS post_path_id_idx ON posts(tree_path, id);

-- flat_sort: sort
CREATE INDEX IF NOT EXISTS post_thread_created_id_idx ON posts(thread, created, id); --checked

-- thread finding
CREATE INDEX IF NOT EXISTS post_thread_id_idx ON posts(thread); -- checked



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
