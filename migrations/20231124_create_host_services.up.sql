-- Create table
CREATE TABLE "host_services"
(
    "id"               serial PRIMARY KEY,
    "host_id"          integer,
    "service_id"       integer,
    "active"           integer   DEFAULT 1,
    "scheduler_number" integer   DEFAULT 3,
    "scheduler_unit"   varchar   DEFAULT 'm',
    "last_check"       timestamp DEFAULT '0001-01-01 00:00:01',
    "created_at"       timestamp DEFAULT now(),
    "updated_at"       timestamp DEFAULT now()
);

-- Create trigger
CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON host_services
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- Add foreign key constraint for host_id
ALTER TABLE "host_services"
    ADD CONSTRAINT fk_host_services_host_id
        FOREIGN KEY ("host_id")
            REFERENCES "hosts" ("id")
            ON DELETE CASCADE
            ON UPDATE CASCADE;

-- Add foreign key constraint for service_id
ALTER TABLE "host_services"
    ADD CONSTRAINT fk_host_services_service_id
        FOREIGN KEY ("service_id")
            REFERENCES "services" ("id")
            ON DELETE CASCADE
            ON UPDATE CASCADE;
