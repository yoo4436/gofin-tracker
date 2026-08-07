-- Migration unit 1: schema_changes
-- Transaction mode: transactional
-- Boundary reason: default

DROP EXTENSION pg_net;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT DELETE, INSERT, SELECT, UPDATE ON TABLES TO anon;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, USAGE ON SEQUENCES TO anon;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON ROUTINES TO anon;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT DELETE, INSERT, SELECT, UPDATE ON TABLES TO authenticated;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, USAGE ON SEQUENCES TO authenticated;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON ROUTINES TO authenticated;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT DELETE, INSERT, SELECT, UPDATE ON TABLES TO service_role;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, USAGE ON SEQUENCES TO service_role;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON ROUTINES TO service_role;

CREATE SEQUENCE public.daily_reports_id_seq AS integer;

CREATE SEQUENCE public.exchange_id_seq AS integer;

CREATE SEQUENCE public.exchange_symbol_id_seq AS integer;

CREATE SEQUENCE public.symbol_id_seq AS integer;

CREATE SEQUENCE public.user_choose_id_seq AS integer;

CREATE SEQUENCE public.users_id_seq AS integer;

CREATE TABLE public.daily_reports (
  id              integer                  DEFAULT nextval('public.daily_reports_id_seq'::regclass) NOT NULL,
  title           character varying(255)   NOT NULL,
  summary         text                     NOT NULL,
  content         text                     NOT NULL,
  cover_image_url text                     DEFAULT ''::text,
  is_premium      boolean                  DEFAULT false,
  published_at    timestamp with time zone DEFAULT now(),
  created_at      timestamp with time zone DEFAULT now(),
  author_id       integer
);

ALTER SEQUENCE public.daily_reports_id_seq OWNED BY public.daily_reports.id;

GRANT ALL ON SEQUENCE public.daily_reports_id_seq TO anon;

GRANT ALL ON SEQUENCE public.daily_reports_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.daily_reports_id_seq TO service_role;

ALTER TABLE public.daily_reports
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.daily_reports
  ADD CONSTRAINT daily_reports_pkey PRIMARY KEY (id);

GRANT ALL ON public.daily_reports TO anon;

GRANT ALL ON public.daily_reports TO authenticated;

GRANT ALL ON public.daily_reports TO service_role;

CREATE INDEX idx_reports_published_at ON public.daily_reports (published_at DESC);

CREATE TABLE public.exchange (
  id     integer               DEFAULT nextval('public.exchange_id_seq'::regclass) NOT NULL,
  name   character varying(50) NOT NULL,
  status character varying(20),
  api    text
);

ALTER SEQUENCE public.exchange_id_seq OWNED BY public.exchange.id;

GRANT ALL ON SEQUENCE public.exchange_id_seq TO anon;

GRANT ALL ON SEQUENCE public.exchange_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.exchange_id_seq TO service_role;

ALTER TABLE public.exchange
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.exchange
  ADD CONSTRAINT exchange_pkey PRIMARY KEY (id);

GRANT ALL ON public.exchange TO anon;

GRANT ALL ON public.exchange TO authenticated;

GRANT ALL ON public.exchange TO service_role;

CREATE TABLE public.exchange_symbol (
  id                integer               DEFAULT nextval('public.exchange_symbol_id_seq'::regclass) NOT NULL,
  exchange_id       integer               NOT NULL,
  symbol_id         integer               NOT NULL,
  trading_pair_code character varying(30)
);

ALTER SEQUENCE public.exchange_symbol_id_seq OWNED BY public.exchange_symbol.id;

GRANT ALL ON SEQUENCE public.exchange_symbol_id_seq TO anon;

GRANT ALL ON SEQUENCE public.exchange_symbol_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.exchange_symbol_id_seq TO service_role;

ALTER TABLE public.exchange_symbol
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.exchange_symbol
  ADD CONSTRAINT exchange_symbol_exchange_id_fkey FOREIGN KEY (exchange_id) REFERENCES public.exchange(id) ON DELETE CASCADE;

ALTER TABLE public.exchange_symbol
  ADD CONSTRAINT exchange_symbol_pkey PRIMARY KEY (id);

ALTER TABLE public.exchange_symbol
  ADD CONSTRAINT unique_exchange_product UNIQUE (exchange_id, symbol_id);

GRANT ALL ON public.exchange_symbol TO anon;

GRANT ALL ON public.exchange_symbol TO authenticated;

GRANT ALL ON public.exchange_symbol TO service_role;

CREATE TABLE public.klines (
  "time"             timestamp with time zone NOT NULL,
  exchange_symbol_id integer                  NOT NULL,
  "interval"         character varying(5)     NOT NULL,
  open_price         numeric(18,8)            NOT NULL,
  high_price         numeric(18,8)            NOT NULL,
  low_price          numeric(18,8)            NOT NULL,
  close_price        numeric(18,8)            NOT NULL,
  volume             numeric(18,4)            NOT NULL,
  created_at         timestamp with time zone DEFAULT now()
);

ALTER TABLE public.klines
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.klines
  ADD CONSTRAINT klines_exchange_symbol_id_fkey FOREIGN KEY (exchange_symbol_id) REFERENCES public.exchange_symbol(id) ON DELETE CASCADE;

ALTER TABLE public.klines
  ADD CONSTRAINT klines_pkey PRIMARY KEY ("time", exchange_symbol_id, "interval");

GRANT ALL ON public.klines TO anon;

GRANT ALL ON public.klines TO authenticated;

GRANT ALL ON public.klines TO service_role;

CREATE TABLE public.report_symbols (
  report_id  integer                  NOT NULL,
  symbol_id  integer                  NOT NULL,
  created_at timestamp with time zone DEFAULT now()
);

ALTER TABLE public.report_symbols
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.report_symbols
  ADD CONSTRAINT report_symbols_pkey PRIMARY KEY (report_id, symbol_id);

ALTER TABLE public.report_symbols
  ADD CONSTRAINT report_symbols_report_id_fkey FOREIGN KEY (report_id) REFERENCES public.daily_reports(id) ON DELETE CASCADE;

GRANT ALL ON public.report_symbols TO anon;

GRANT ALL ON public.report_symbols TO authenticated;

GRANT ALL ON public.report_symbols TO service_role;

CREATE TABLE public.symbol (
  id          integer               DEFAULT nextval('public.symbol_id_seq'::regclass) NOT NULL,
  symbol_code character varying(30) NOT NULL,
  name        character varying(50) NOT NULL,
  market_type character varying(20) NOT NULL
);

ALTER SEQUENCE public.symbol_id_seq OWNED BY public.symbol.id;

GRANT ALL ON SEQUENCE public.symbol_id_seq TO anon;

GRANT ALL ON SEQUENCE public.symbol_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.symbol_id_seq TO service_role;

ALTER TABLE public.symbol
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.symbol
  ADD CONSTRAINT symbol_pkey PRIMARY KEY (id);

ALTER TABLE public.exchange_symbol
  ADD CONSTRAINT exchange_symbol_symbol_id_fkey FOREIGN KEY (symbol_id) REFERENCES public.symbol(id) ON DELETE CASCADE;

ALTER TABLE public.report_symbols
  ADD CONSTRAINT report_symbols_symbol_id_fkey FOREIGN KEY (symbol_id) REFERENCES public.symbol(id) ON DELETE CASCADE;

ALTER TABLE public.symbol
  ADD CONSTRAINT symbol_symbol_code_key UNIQUE (symbol_code);

GRANT ALL ON public.symbol TO anon;

GRANT ALL ON public.symbol TO authenticated;

GRANT ALL ON public.symbol TO service_role;

CREATE TABLE public.user_choose (
  id                 integer DEFAULT nextval('public.user_choose_id_seq'::regclass) NOT NULL,
  user_id            integer NOT NULL,
  exchange_symbol_id integer NOT NULL,
  sort_order         integer DEFAULT 0
);

ALTER SEQUENCE public.user_choose_id_seq OWNED BY public.user_choose.id;

GRANT ALL ON SEQUENCE public.user_choose_id_seq TO anon;

GRANT ALL ON SEQUENCE public.user_choose_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.user_choose_id_seq TO service_role;

ALTER TABLE public.user_choose
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.user_choose
  ADD CONSTRAINT unique_user_favorite UNIQUE (user_id, exchange_symbol_id);

ALTER TABLE public.user_choose
  ADD CONSTRAINT user_choose_exchange_symbol_id_fkey FOREIGN KEY (exchange_symbol_id) REFERENCES public.exchange_symbol(id) ON DELETE CASCADE;

ALTER TABLE public.user_choose
  ADD CONSTRAINT user_choose_pkey PRIMARY KEY (id);

GRANT ALL ON public.user_choose TO anon;

GRANT ALL ON public.user_choose TO authenticated;

GRANT ALL ON public.user_choose TO service_role;

CREATE TABLE public.users (
  id        integer                  DEFAULT nextval('public.users_id_seq'::regclass) NOT NULL,
  name      character varying(100)   NOT NULL,
  pwd       character varying(255)   NOT NULL,
  email     character varying(100)   NOT NULL,
  tel       character varying(20),
  create_at timestamp with time zone DEFAULT now()
);

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

GRANT ALL ON SEQUENCE public.users_id_seq TO anon;

GRANT ALL ON SEQUENCE public.users_id_seq TO authenticated;

GRANT ALL ON SEQUENCE public.users_id_seq TO service_role;

ALTER TABLE public.users
  ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.users
  ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE public.users
  ADD CONSTRAINT users_name_key UNIQUE (name);

ALTER TABLE public.users
  ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE public.daily_reports
  ADD CONSTRAINT daily_reports_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id);

ALTER TABLE public.user_choose
  ADD CONSTRAINT user_choose_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

GRANT ALL ON public.users TO anon;

GRANT ALL ON public.users TO authenticated;

GRANT ALL ON public.users TO service_role;
