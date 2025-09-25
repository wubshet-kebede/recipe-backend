alter table "public"."recipes" alter column "image_url" drop not null;
alter table "public"."recipes" add column "image_url" text;
