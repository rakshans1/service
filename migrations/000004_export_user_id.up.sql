BEGIN;
ALTER TABLE products
	ADD COLUMN user_id UUID DEFAULT '00000000-0000-0000-0000-000000000000';
END;