alter table "public"."ingredients" add constraint "ingredients_recipe_id_key" unique ("recipe_id");
alter table "public"."ingredients" alter column "recipe_id" set default gen_random_uuid();
