drop table users if exists;
drop table todos if exists;

create table todos (
  id serial primary key,
  created_at timestamp not null,
  content text,
  active boolean default false
);
