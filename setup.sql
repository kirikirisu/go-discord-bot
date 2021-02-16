create table users (
  id serial primary key,
  name varcha(255),
)

create table todos (
  id serial primary key,
  user_id integer references users(id),
  created_at timestamp not null,
  content text,
)
