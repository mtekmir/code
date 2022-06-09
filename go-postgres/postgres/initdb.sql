create table if not exists brands(
  id bigserial unique primary key,
  name varchar not null
);

create table if not exists variations(
  id bigserial unique primary key,
  name varchar not null
);

create table if not exists products(
  id bigserial unique primary key,
  created_at timestamptz default now(),
  name varchar not null,
  price int not null,
  brand_id bigint references brands(id)
);

create table if not exists product_variations(
  id bigserial unique primary key,
  quantity int not null,
  product_id bigint not null references products(id),
  variation_id bigint not null references variations(id)
);
