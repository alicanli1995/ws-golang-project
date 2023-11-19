-- Create table
CREATE TABLE "preferences"
(
    "id"         serial PRIMARY KEY,
    "name"       varchar(255),
    "preference" text,
    "created_at" timestamp,
    "updated_at" timestamp
);

-- Set default values for created_at and updated_at
ALTER TABLE "preferences"
    ALTER COLUMN "created_at" SET DEFAULT now();
ALTER TABLE "preferences"
    ALTER COLUMN "updated_at" SET DEFAULT now();

-- Create trigger
CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON preferences
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
