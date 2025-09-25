alter table "public"."users" alter column "profile_picture_url" drop not null;
alter table "public"."users" add column "profile_picture_url" text;
