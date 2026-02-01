CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(300),
    value DECIMAL NOT NULL,
    stock INTEGER NOT NULL
);