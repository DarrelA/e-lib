-- Create the User table
CREATE TABLE IF NOT EXISTS Users(
  id bigint PRIMARY KEY,
  name varchar(255) NOT NULL,
  email varchar(255) UNIQUE NOT NULL, -- Email should be unique
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create the Book table
CREATE TABLE IF NOT EXISTS Books(
  uuid uuid PRIMARY KEY,
  title varchar(255) UNIQUE NOT NULL, -- Title should be unique
  available_copies integer NOT NULL DEFAULT 0 CHECK (available_copies >= 0), -- Ensure non-negative
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create the Loan table
CREATE TABLE IF NOT EXISTS Loans(
  uuid uuid PRIMARY KEY,
  user_id bigint NOT NULL,
  book_uuid uuid NOT NULL,
  name_of_borrower varchar(255) NOT NULL,
  loan_date timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  return_date timestamp with time zone NOT NULL,
  is_returned boolean NOT NULL DEFAULT FALSE,
  FOREIGN KEY (user_id) REFERENCES Users(id),
  FOREIGN KEY (book_uuid) REFERENCES Books(uuid)
);

-- Indexes for performance
-- For searching loans by user
CREATE INDEX IF NOT EXISTS idx_loans_user_id ON Loans(user_id);

-- For searching loans by book
CREATE INDEX IF NOT EXISTS idx_loans_book_uuid ON Loans(book_uuid);

-- For efficiently finding overdue/active loans
CREATE INDEX IF NOT EXISTS idx_loans_is_returned ON Loans(is_returned);

