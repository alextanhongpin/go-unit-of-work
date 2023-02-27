insert into users(email)
values ($1)
returning id, email;
