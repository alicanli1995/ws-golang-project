-- Create table
CREATE TABLE "events" (
    "id" serial PRIMARY KEY,
    "event_type" varchar(255),
    "host_service_id" integer,
    "host_id" integer,
    "service_name" varchar(255),
    "host_name" varchar(255),
    "message" varchar(255),
    "created_at" timestamp NOT NULL DEFAULT NOW(),
    "updated_at" timestamp NOT NULL DEFAULT NOW()
);

-- Create trigger
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
