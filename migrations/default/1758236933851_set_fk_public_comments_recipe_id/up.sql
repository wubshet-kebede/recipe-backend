alter table "public"."comments" drop constraint "comments_recipe_id_fkey",
  add constraint "comments_recipe_id_fkey"
  foreign key ("recipe_id")
  references "public"."recipes"
  ("id") on update restrict on delete cascade;
