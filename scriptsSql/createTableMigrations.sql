CREATE TABLE IF NOT EXISTS public.t_migrations
(
    id serial primary key,
    migration_name text not null
)