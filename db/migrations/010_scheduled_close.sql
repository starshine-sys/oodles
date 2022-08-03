-- 2021-08-03
-- Add scheduled close

-- +migrate Up

alter table applications add column scheduled_close_id bigint;
