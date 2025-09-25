ALTER TABLE "public"."ingredients" ALTER COLUMN "recipe_id" drop default;
alter table "public"."ingredients" drop constraint "ingredients_recipe_id_key";
