-- 2021-01-31
-- Add reminders, and user config table

-- +migrate Up

create or replace view reminders as
    select id, expires,
    (data->'user_id'->>0)::bigint as user_id,
    (data->'guild_id'->>0)::bigint as guild_id, 
    (data->'channel_id'->>0)::bigint as channel_id,
    (data->'message_id'->>0)::bigint as message_id,
    data->'text'->>0 as reminder_text,
    (data->'set_time'->>0)::timestamp as set_time
    from public.scheduled_events where event_type = 'reminders.reminder';

create extension hstore;

create table users (
    id      bigint  primary key,
    config  hstore
);
