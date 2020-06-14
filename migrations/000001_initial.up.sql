BEGIN;
CREATE TABLE products (
	product_id UUID,
	name TEXT,
	cost INT,
	quantity INT,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY kEY (product_id)
);
END;
