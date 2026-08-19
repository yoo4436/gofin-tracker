CREATE INDEX IF NOT EXISTS idx_klines_exchange_symbol_interval_time
  ON public.klines (exchange_symbol_id, interval, "time" DESC);

-- Future objects should be private until a migration grants the minimum
-- required privilege and enables the intended RLS policy explicitly.
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
  REVOKE ALL ON TABLES FROM anon, authenticated;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
  REVOKE ALL ON SEQUENCES FROM anon, authenticated;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
  REVOKE ALL ON ROUTINES FROM anon, authenticated;
