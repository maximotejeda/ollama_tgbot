PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS models (
       id INTEGER PRIMARY KEY NOT NULL,
       model_name TEXT NOT NULL,
       model_tag TEXT NOT NULL,
       created TEXT NOT NULL,
       edited TEXT NOT NULL,
       deleted TEXT,
       UNIQUE(model_name, model_tag)
);

CREATE TABLE IF NOT EXISTS users (
       id INTEGER PRIMARY KEY NOT NULL,
       t_id INTEGER UNIQUE,
       username TEXT NOT NULL,
       first_name TEXT,
       last_name TEXT,
       auth TEXT,
       model_id INTEGER,
       mode TEXT,
       created INTEGER NOT NULL,
       edited TEXT NOT NULL,
       deleted TEXT,
       FOREIGN KEY(model_id) REFERENCES models(id) ON DELETE CASCADE ON UPDATE CASCADE,
       UNIQUE(username, first_name, last_name)
);


CREATE TABLE IF NOT EXISTS history (
       id INTEGER PRIMARY KEY NOT NULL,
       user_id INTEGER NOT NULL,
       query_mode TEXT, -- chat or generatedepends of the type of access if empty chat
       conversation TEXT,  -- group of prompts sent before will be limited to 10
       model_id INTEGER,
       created TEXT NOT NULL,
       edited TEXT NOT NULL, -- last prompt if more than x time delete messages
       deleted TEXT,
       FOREIGN KEY(model_id) REFERENCES models(id) ON DELETE CASCADE ON UPDATE CASCADE,
       FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

