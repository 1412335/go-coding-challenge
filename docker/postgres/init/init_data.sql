-- Create the primary key generator

create sequence public.table_id_seq;

CREATE OR REPLACE FUNCTION public.next_id(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    shard_id int := 5;
BEGIN
    SELECT nextval('public.table_id_seq') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id << 10);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

-- Create table for users
DROP TABLE IF EXISTS "public"."users";
CREATE TABLE "public"."users" (
    id  bigint NOT NULL DEFAULT next_id(),
    email varchar(255) NOT NULL,
    password varchar(255)  NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- Records of users
INSERT INTO "public"."users" (email, password)
VALUES ('abc@gmail.com', '$2a$10$Ir83oelk9psNRBosHTgICe2Woe5FMXNbRLBQSS/Jwq.T9i5Gcvlgu');


-- -- Create table for accounts
-- CREATE TABLE "public"."accounts" (
--     id  varchar(50)  NOT NULL,
--     created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     PRIMARY KEY (id)
-- );

-- -- Insert sample promotion data
-- insert into promotion (id, , validfrom, validthru, customerid)
-- VALUES ('Promo1', 'Promo1 Desc', TIMESTAMP '2000-01-01 00:00:00', TIMESTAMP '2200-01-01 00:00:00','ducksrus'),
--        ('Promo2', 'Promo2 Desc', TIMESTAMP '2000-01-01 00:00:00', TIMESTAMP '2200-01-01 00:00:00','patoloco'),
--        ('Promo3', 'Promo3 Desc', TIMESTAMP '2000-01-01 00:00:00', TIMESTAMP '2200-01-01 00:00:00','ducksrus')