DROP TABLE IF EXISTS Actual;

DROP TABLE IF EXISTS Expected;

CREATE TABLE Expected(
  id serial PRIMARY KEY,
  method varchar(255) NOT NULL,
  url_path varchar(255) NOT NULL,
  status_code int NOT NULL,
  res_body_contains text
);

CREATE TABLE Actual(
  id serial PRIMARY KEY,
  expected_id int NOT NULL,
  status_code int NOT NULL,
  req_url_query_string text,
  req_body text,
  res_body text,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (expected_id) REFERENCES Expected(id)
);

