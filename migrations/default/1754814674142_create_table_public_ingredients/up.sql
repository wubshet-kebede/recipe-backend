CREATE TABLE "public"."ingredients" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "recipe_id" uuid NOT NULL DEFAULT gen_random_uuid(), "name" text NOT NULL, "quantity" text NOT NULL, PRIMARY KEY ("id") , FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes"("id") ON UPDATE restrict ON DELETE cascade, UNIQUE ("id"), UNIQUE ("recipe_id"));
CREATE EXTENSION IF NOT EXISTS pgcrypto;
