
CREATE TABLE public.clients (
    id bigint NOT NULL,
    name text,
    email text,
    company text,
    phone text,
    notes text,
    owner_id bigint,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.clients_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.clients_id_seq OWNED BY public.clients.id;

CREATE TABLE public.directory_deletion_state (
    id integer NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.documents (
    id bigint NOT NULL,
    name text,
    status text DEFAULT 'draft'::text,
    file_name text,
    owner_id bigint,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    storage_path text,
    original_hash text,
    signed_hash text,
    sequential boolean DEFAULT false,
    space_id bigint,
    client_id bigint,
    external_ref character varying(200) DEFAULT ''::character varying NOT NULL,
    nuage_space_facile_id character varying(40) DEFAULT ''::character varying NOT NULL,
    nuage_linked_to character varying(200) DEFAULT ''::character varying NOT NULL,
    nuage_published_at timestamp with time zone
);

CREATE SEQUENCE public.documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.documents_id_seq OWNED BY public.documents.id;

CREATE TABLE public.fields (
    id bigint NOT NULL,
    document_id bigint,
    signer_id bigint,
    field_type text,
    page bigint,
    x numeric,
    y numeric,
    width numeric,
    height numeric,
    required boolean,
    value text,
    created_at timestamp with time zone DEFAULT now(),
    label text
);

CREATE SEQUENCE public.fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.fields_id_seq OWNED BY public.fields.id;

CREATE TABLE public.mcp_actions (
    id text NOT NULL,
    user_id text NOT NULL,
    kind text NOT NULL,
    target text NOT NULL,
    digest text NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    client_id text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    resolved_at timestamp with time zone,
    outcome text DEFAULT ''::text NOT NULL
);

CREATE TABLE public.porte_identities (
    user_id bigint NOT NULL,
    provider text NOT NULL,
    subject text NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL,
    access_token text DEFAULT ''::text NOT NULL,
    refresh_token text DEFAULT ''::text NOT NULL,
    token_expiry timestamp with time zone,
    roles jsonb,
    roles_synced_at timestamp with time zone,
    synced_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE public.porte_magic_links (
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);

CREATE TABLE public.porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE public.porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.porte_sessions_id_seq OWNED BY public.porte_sessions.id;

CREATE TABLE public.signers (
    id bigint NOT NULL,
    document_id bigint,
    name text,
    email text,
    role text DEFAULT 'signer'::text,
    status text DEFAULT 'pending'::text,
    token text,
    order_num bigint DEFAULT 0,
    signed_at timestamp with time zone,
    ip_address text,
    user_agent text,
    created_at timestamp with time zone DEFAULT now(),
    last_reminded_at timestamp with time zone,
    viewed_at timestamp with time zone,
    email_opened_at timestamp with time zone
);

CREATE SEQUENCE public.signers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.signers_id_seq OWNED BY public.signers.id;

CREATE TABLE public.smtp_configs (
    id bigint NOT NULL,
    owner_id bigint,
    host text,
    port bigint,
    username text,
    password text,
    from_email text,
    from_name text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    space_id bigint
);

CREATE SEQUENCE public.smtp_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.smtp_configs_id_seq OWNED BY public.smtp_configs.id;

CREATE TABLE public.space_members (
    id bigint NOT NULL,
    space_id bigint,
    user_id bigint,
    role text DEFAULT 'member'::text,
    joined_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.space_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.space_members_id_seq OWNED BY public.space_members.id;

CREATE TABLE public.spaces (
    id bigint NOT NULL,
    name text,
    description text,
    owner_id bigint,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    facile_id text,
    registre_space_id text
);

CREATE SEQUENCE public.spaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.spaces_id_seq OWNED BY public.spaces.id;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    name text,
    password_hash text DEFAULT ''::text,
    created_at timestamp with time zone DEFAULT now(),
    reminder_interval_days bigint DEFAULT 3,
    avatar_url text DEFAULT ''::text,
    avatar_source text DEFAULT ''::text,
    oidc_picture_url text,
    oidc_access_token text DEFAULT ''::text,
    oidc_refresh_token text DEFAULT ''::text,
    oidc_token_expiry timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    profile_synced_at timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    oidc_subject text,
    avatar_upload_path text DEFAULT ''::text,
    locale text DEFAULT 'en'::text NOT NULL
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE public.webhooks (
    id bigint NOT NULL,
    owner_id bigint,
    url text,
    secret text,
    enabled boolean DEFAULT true,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    space_id bigint
);

CREATE SEQUENCE public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;

ALTER TABLE ONLY public.clients ALTER COLUMN id SET DEFAULT nextval('public.clients_id_seq'::regclass);

ALTER TABLE ONLY public.documents ALTER COLUMN id SET DEFAULT nextval('public.documents_id_seq'::regclass);

ALTER TABLE ONLY public.fields ALTER COLUMN id SET DEFAULT nextval('public.fields_id_seq'::regclass);

ALTER TABLE ONLY public.porte_sessions ALTER COLUMN id SET DEFAULT nextval('public.porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY public.signers ALTER COLUMN id SET DEFAULT nextval('public.signers_id_seq'::regclass);

ALTER TABLE ONLY public.smtp_configs ALTER COLUMN id SET DEFAULT nextval('public.smtp_configs_id_seq'::regclass);

ALTER TABLE ONLY public.space_members ALTER COLUMN id SET DEFAULT nextval('public.space_members_id_seq'::regclass);

ALTER TABLE ONLY public.spaces ALTER COLUMN id SET DEFAULT nextval('public.spaces_id_seq'::regclass);

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.webhooks ALTER COLUMN id SET DEFAULT nextval('public.webhooks_id_seq'::regclass);

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fields_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY public.porte_magic_links
    ADD CONSTRAINT porte_magic_links_pkey PRIMARY KEY (token_hash);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY public.signers
    ADD CONSTRAINT signers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.smtp_configs
    ADD CONSTRAINT smtp_configs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.space_members
    ADD CONSTRAINT space_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spaces
    ADD CONSTRAINT spaces_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);

CREATE INDEX idx_clients_owner_id ON public.clients USING btree (owner_id);

CREATE INDEX idx_documents_client_id ON public.documents USING btree (client_id);

CREATE INDEX idx_documents_original_hash ON public.documents USING btree (original_hash);

CREATE UNIQUE INDEX idx_documents_owner_external_ref ON public.documents USING btree (owner_id, external_ref) WHERE ((external_ref)::text <> ''::text);

CREATE INDEX idx_documents_owner_id ON public.documents USING btree (owner_id);

CREATE INDEX idx_documents_signed_hash ON public.documents USING btree (signed_hash);

CREATE INDEX idx_documents_space_id ON public.documents USING btree (space_id);

CREATE INDEX idx_documents_status ON public.documents USING btree (status);

CREATE INDEX idx_fields_document_id ON public.fields USING btree (document_id);

CREATE INDEX idx_fields_signer_id ON public.fields USING btree (signer_id);

CREATE INDEX idx_signers_document_id ON public.signers USING btree (document_id);

CREATE INDEX idx_signers_token ON public.signers USING btree (token);

CREATE UNIQUE INDEX idx_smtp_configs_owner_id ON public.smtp_configs USING btree (owner_id);

CREATE INDEX idx_smtp_configs_space_id ON public.smtp_configs USING btree (space_id);

CREATE INDEX idx_space_members_space_id ON public.space_members USING btree (space_id);

CREATE UNIQUE INDEX idx_space_members_space_user ON public.space_members USING btree (space_id, user_id);

CREATE INDEX idx_space_members_user_id ON public.space_members USING btree (user_id);

CREATE UNIQUE INDEX idx_spaces_facile_id ON public.spaces USING btree (facile_id);

CREATE INDEX idx_spaces_owner_id ON public.spaces USING btree (owner_id);

CREATE UNIQUE INDEX idx_spaces_registre_space_id ON public.spaces USING btree (registre_space_id);

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON public.users USING btree (oidc_subject);

CREATE INDEX idx_webhooks_owner_id ON public.webhooks USING btree (owner_id);

CREATE INDEX idx_webhooks_space_id ON public.webhooks USING btree (space_id);

CREATE INDEX mcp_actions_open_idx ON public.mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON public.porte_identities USING btree (user_id);

CREATE INDEX porte_sessions_expiry_idx ON public.porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON public.porte_sessions USING btree (user_id);

ALTER TABLE ONLY public.space_members
    ADD CONSTRAINT fk_space_members_space FOREIGN KEY (space_id) REFERENCES public.spaces(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.space_members
    ADD CONSTRAINT fk_space_members_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_magic_links
    ADD CONSTRAINT porte_magic_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

