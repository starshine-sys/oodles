-- +migrate Up

create table guilds (
    id      bigint  not null    primary key,

    config      jsonb   not null    default '{}'::jsonb, -- json configuration
    commands    jsonb   not null    default '{}'::jsonb, -- command overrides
    perms       jsonb   not null    default '{}'::jsonb -- user/role permissions
);

create table application_tracks (
    id          serial  primary key,
    name        text    not null,
    description text    not null,
    emoji       text    not null unique
);

create table app_questions (
    track_id    bigint  not null    references application_tracks (id) on delete cascade,
    id          serial  primary key,
    question    text    not null,
    long_answer boolean not null    default false
);

create table applications (
    id          serial  primary key,
    user_id     bigint  not null,
    channel_id  bigint  not null,

    -- this isn't perfect, but i'd rather have deleted/edited application tracks show up as "unknown" than delete old logs entirely
    track_id    bigint  references application_tracks (id) on delete set null,

    -- not question ID, but index
    question    int     not null    default 0,

    -- when the application was opened
    opened  timestamp   not null    default (current_timestamp at time zone 'utc'),
    -- whether the application has been completed (finished the entire track)
    completed   bool    not null    default false,
    -- whether the user was verified
    verified    bool,
    -- if denied, why the user was denied
    deny_reason text,
    -- if verified or denied, the moderator who verified/denied the user
    moderator   bigint,
    -- whether the application is closed (user verified/denied, channel deleted)
    closed      bool        not null    default false,
    closed_time timestamp,

    -- for linking to transcripts
    transcript_channel  bigint,
    transcript_message  bigint
);

create table app_responses (
    application_id  bigint  not null    references applications (id) on delete cascade,
    message_id      bigint  not null    primary key,
    user_id         bigint  not null,
    username        text    not null,
    discriminator   text    not null,
    content         text    not null,

    from_bot    bool    not null    default false,
    from_staff  bool    not null    default false
);

create index applications_user_idx on applications (user_id);
create index responses_app_idx on app_responses (application_id);
