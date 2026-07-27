ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));
