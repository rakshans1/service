BEGIN;
ALTER TABLE products
	DROP COLUMN user_id;
END;