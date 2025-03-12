DROP TABLE IF EXISTS Actual;

CREATE TABLE Actual(
  id serial PRIMARY KEY,
  expected_id int,
  status_code int NOT NULL,
  req_url_query_string text,
  req_body text,
  res_body text,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (expected_id) REFERENCES Expected(id)
);

