BEGIN;

CREATE TABLE currencies (
    code VARCHAR(3) PRIMARY KEY,
    name TEXT NOT NULL,
    decimal_places SMALLINT NOT NULL,

    CONSTRAINT currencies_code_format
        CHECK (code ~ '^[A-Z]{3}$'),
    CONSTRAINT currencies_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT currencies_decimal_places_range
        CHECK (decimal_places BETWEEN 0 AND 4)
);

INSERT INTO currencies (code, name, decimal_places)
VALUES
    ('EUR', 'Euro', 2),
    ('GBP', 'Pound Sterling', 2),
    ('JPY', 'Japanese Yen', 0),
    ('TRY', 'Turkish Lira', 2),
    ('USD', 'US Dollar', 2);

COMMIT;
