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
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- Records of users
INSERT INTO "public"."users" (id, email, password)
VALUES (1, 'abc@gmail.com', '$2a$10$Ir83oelk9psNRBosHTgICe2Woe5FMXNbRLBQSS/Jwq.T9i5Gcvlgu');


-- Create table for accounts
DROP TABLE IF EXISTS "public"."accounts";
CREATE TABLE "public"."accounts" (
    id  bigint NOT NULL DEFAULT next_id(),
    user_id bigint NOT NULL,
    name varchar(255),
    bank varchar(10),
    balance numeric NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- Insert sample account data
insert into "public"."accounts" (id, user_id, name, bank, balance)
VALUES (1, 1, '1.VIB', 'VIB', 1000000);


-- Create table for accounts
DROP TABLE IF EXISTS "public"."transactions";
CREATE TABLE "public"."transactions" (
    id  bigint NOT NULL DEFAULT next_id(),
    account_id bigint NOT NULL,
    transaction_type varchar(10),
    amount numeric NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- Insert sample account data
insert into "public"."transactions" (id, account_id, transaction_type, amount)
VALUES (1, 1, 'DEPOSIT', 1000000),
        (2, 1, 'WITHDRAW', 100000);