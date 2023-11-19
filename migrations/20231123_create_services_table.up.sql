-- Create table
CREATE TABLE "services"
(
    "id"           serial PRIMARY KEY,
    "service_name" varchar(255),
    "active"       integer DEFAULT 1,
    "icon"         varchar(255),
    "created_at"   timestamp DEFAULT now(),
    "updated_at"   timestamp DEFAULT now()
);

-- Create trigger
CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON services
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
