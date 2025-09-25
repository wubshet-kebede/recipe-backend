CREATE TABLE IF NOT EXISTS public.order_items (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL,
    recipe_id uuid NOT NULL,
    quantity integer NOT NULL,
    price_at_purchase numeric NOT NULL,
    recipe_name text NOT NULL,
    recipe_image_url text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),


    CONSTRAINT order_items_pkey PRIMARY KEY (id),


    CONSTRAINT fk_order_id FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE,

 
    CONSTRAINT fk_recipe_id FOREIGN KEY (recipe_id) REFERENCES public.recipes(id) ON DELETE CASCADE
);


CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON public.order_items USING btree (order_id);


CREATE TRIGGER update_order_items_updated_at BEFORE UPDATE
    ON public.order_items FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
