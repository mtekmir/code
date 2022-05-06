create table if not exists variations(
  id bigserial unique primary key,
  name varchar not null
);

create table if not exists products(
  id bigserial unique primary key,
  barcode varchar not null,
  name varchar not null,
  price numeric(10,2),
  quantity int not null,
  variation_id bigint references variations(id)
);