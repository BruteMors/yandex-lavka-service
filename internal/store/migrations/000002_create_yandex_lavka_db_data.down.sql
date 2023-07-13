DELETE FROM public.courier_types WHERE courier_type IN ('FOOT', 'BIKE', 'AUTO');
TRUNCATE public.courier_types RESTART IDENTITY CASCADE;