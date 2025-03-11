DROP TABLE IF EXISTS actual;

DROP TABLE IF EXISTS expected;

CREATE TABLE expected(
  id int PRIMARY KEY,
  method varchar(255) NOT NULL,
  url varchar(255) NOT NULL,
  statusCode int NOT NULL,
  resBodyContains text
);

CREATE TABLE actual(
  id int PRIMARY KEY,
  method varchar(255) NOT NULL,
  expectedId int NOT NULL,
  statusCode int NOT NULL,
  reqBody text,
  createdAt timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (expectedId) REFERENCES expected(id)
);

