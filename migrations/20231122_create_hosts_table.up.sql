-- Create table
CREATE TABLE "hosts"
(
    "id"             serial PRIMARY KEY,
    "host_name"      varchar(255),
    "canonical_name" varchar(255),
    "url"            varchar(255),
    "ip"             varchar(255),
    "ipv6"           varchar(255),
    "location"       varchar(255),
    "os"             varchar(255),
    "active"         integer DEFAULT 1,
    "created_at"     timestamp DEFAULT now(),
    "updated_at"     timestamp DEFAULT now()
);

-- Create trigger
CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON hosts
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
