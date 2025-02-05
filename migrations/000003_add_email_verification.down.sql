ALTER TABLE users
  DROP COLUMN email_verified,
  DROP COLUMN verification_token,
  DROP COLUMN verification_expires_at;
