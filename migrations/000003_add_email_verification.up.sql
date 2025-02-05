ALTER TABLE users
  ADD COLUMN email_verified BOOLEAN DEFAULT FALSE,
  ADD COLUMN verification_token VARCHAR(64),
  ADD COLUMN verification_expires_at TIMESTAMP WITH TIME ZONE;
