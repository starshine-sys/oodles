-- 2021-01-31
-- Add scheduler

-- +migrate Up

create table scheduled_events (
    id          serial      primary key,
    event_type  text        not null,
    expires     timestamp   not null,
    data        jsonb       not null
);

create index scheduled_events_expires_idx on scheduled_events (expires);
create index scheduled_events_data_idx on scheduled_events using GIN (data);

alter table applications add column scheduled_event_id int;
