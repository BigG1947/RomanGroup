CREATE TABLE IF NOT EXISTS news
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT,
    title       VARCHAR(512) NOT NULL UNIQUE,
    description TEXT         NOT NULL,
    text        TEXT         NOT NULL,
    images      VARCHAR(512) NOT NULL,
    date        TIMESTAMP    NOT NULL DEFAULT NOW(),
    count_see   INTEGER      NOT NULL DEFAULT 0
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";

CREATE TABLE IF NOT EXISTS feedback
(
    id       INTEGER PRIMARY KEY AUTO_INCREMENT,
    name     VARCHAR(256) NOT NULL,
    email    VARCHAR(256) NOT NULL,
    theme    VARCHAR(256) NOT NULL,
    text     TEXT         NOT NULL,
    date     TIMESTAMP    NOT NULL DEFAULT NOW(),
    is_check BOOLEAN      NOT NULL DEFAULT FALSE
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";


CREATE TABLE IF NOT EXISTS project
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(512) NOT NULL,
    description TEXT         NOT NULL,
    images      VARCHAR(512) NOT NULL,
    is_favorite INTEGER      NOT NULL DEFAULT 0,
    video1      VARCHAR(256),
    video2      VARCHAR(256),
    video3      VARCHAR(256),
    date        TIMESTAMP    NOT NULL DEFAULT NOW()
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";

CREATE TABLE IF NOT EXISTS project_photo
(
    id         INTEGER PRIMARY KEY AUTO_INCREMENT,
    src        VARCHAR(512) NOT NULL,
    date       TIMESTAMP    NOT NULL DEFAULT NOW(),
    id_project INTEGER      NOT NULL,
    FOREIGN KEY (id_project) REFERENCES project (id)
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";

CREATE TABLE IF NOT EXISTS admin
(
    id       INTEGER PRIMARY KEY AUTO_INCREMENT,
    login    VARCHAR(128) NOT NULL,
    password VARCHAR(256) NOT NULL
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";

INSERT INTO admin(id, login, password)
VALUES (1, "admin", "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918");

CREATE TABLE IF NOT EXISTS social
(
    name VARCHAR(128) PRIMARY KEY NOT NULL,
    url  VARCHAR(256)
) ENGINE = InnoDB,
  CHARSET = "utf8mb4";

INSERT INTO social(name, url)
VALUES ("facebook", null);
INSERT INTO social(name, url)
VALUES ("viber", null);
INSERT INTO social(name, url)
VALUES ("telegram", null);
INSERT INTO social(name, url)
VALUES ("youtube", null);
