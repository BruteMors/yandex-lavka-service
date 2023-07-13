CREATE TABLE public.couriers (
    courier_id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    courier_type_id integer NOT NULL,
    PRIMARY KEY (courier_id)
);

CREATE INDEX ON public.couriers
    (courier_type_id);


CREATE TABLE public.courier_types (
    courier_type_id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    courier_type varchar(32) NOT NULL,
    PRIMARY KEY (courier_type_id)
);

ALTER TABLE public.courier_types
    ADD UNIQUE (courier_type);


CREATE TABLE public.orders (
    order_id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    courier_id integer,
    region integer NOT NULL,
    weight numeric NOT NULL,
    cost integer NOT NULL,
    completed_time timestamp with time zone,
    PRIMARY KEY (order_id)
);

CREATE INDEX ON public.orders
    (courier_id);


CREATE TABLE public.delivery_hours (
    delivery_hours_id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    order_id integer NOT NULL,
    delivery_interval char(11) NOT NULL,
    PRIMARY KEY (delivery_hours_id)
);

CREATE INDEX ON public.delivery_hours
    (order_id);


CREATE TABLE public.working_hours (
    working_hours_id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    courier_id integer NOT NULL,
    working_interval char(11) NOT NULL,
    PRIMARY KEY (working_hours_id)
);

CREATE INDEX ON public.working_hours
    (courier_id);


CREATE TABLE public.couriers_to_regions (
    courier_to_region INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    courier_id integer NOT NULL,
    region integer NOT NULL,
    PRIMARY KEY (courier_to_region)
);

CREATE INDEX ON public.couriers_to_regions
    (courier_id);


ALTER TABLE public.couriers ADD CONSTRAINT FK_couriers__courier_type_id FOREIGN KEY (courier_type_id) REFERENCES public.courier_types(courier_type_id);
ALTER TABLE public.orders ADD CONSTRAINT FK_orders__courier_id FOREIGN KEY (courier_id) REFERENCES public.couriers(courier_id);
ALTER TABLE public.delivery_hours ADD CONSTRAINT FK_delivery_hours__order_id FOREIGN KEY (order_id) REFERENCES public.orders(order_id);
ALTER TABLE public.working_hours ADD CONSTRAINT FK_working_hours__courier_id FOREIGN KEY (courier_id) REFERENCES public.couriers(courier_id);
ALTER TABLE public.couriers_to_regions ADD CONSTRAINT FK_couriers_to_regions__courier_id FOREIGN KEY (courier_id) REFERENCES public.couriers(courier_id);