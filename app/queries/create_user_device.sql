insert into user_devices(user_id, device_id)
values ($1, $2)
returning id, user_id, device_id;
