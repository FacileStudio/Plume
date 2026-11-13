-- +goose Up
-- baseline: the production schema as GORM AutoMigrate and schemas.MigrateWithIssuer built it, captured 2026-11-13 with pg_dump --schema-only --no-owner --no-privileges --no-comments --no-tablespaces --schema=public

CREATE TABLE clients (
    id bigint NOT NULL,
    name text,
    email text,
    company text,
    phone text,
    notes text,
    owner_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE clients_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE clients_id_seq OWNED BY clients.id;

CREATE TABLE directory_deletion_state (
    id integer NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE documents (
    id bigint NOT NULL,
    name text,
    status text DEFAULT 'draft'::text,
    file_name text,
    owner_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
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

CREATE SEQUENCE documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE documents_id_seq OWNED BY documents.id;

CREATE TABLE fields (
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
    created_at timestamp with time zone,
    label text
);

CREATE SEQUENCE fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE fields_id_seq OWNED BY fields.id;

CREATE TABLE mcp_actions (
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

CREATE TABLE porte_identities (
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

CREATE TABLE porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE porte_magic_links (
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);

CREATE TABLE porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE porte_sessions_id_seq OWNED BY porte_sessions.id;

CREATE TABLE signers (
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
    created_at timestamp with time zone,
    last_reminded_at timestamp with time zone,
    viewed_at timestamp with time zone,
    email_opened_at timestamp with time zone
);

CREATE SEQUENCE signers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE signers_id_seq OWNED BY signers.id;

CREATE TABLE smtp_configs (
    id bigint NOT NULL,
    owner_id bigint,
    host text,
    port bigint,
    username text,
    password text,
    from_email text,
    from_name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    space_id bigint
);

CREATE SEQUENCE smtp_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE smtp_configs_id_seq OWNED BY smtp_configs.id;

CREATE TABLE space_members (
    id bigint NOT NULL,
    space_id bigint,
    user_id bigint,
    role text DEFAULT 'member'::text,
    joined_at timestamp with time zone
);

CREATE SEQUENCE space_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE space_members_id_seq OWNED BY space_members.id;

CREATE TABLE spaces (
    id bigint NOT NULL,
    name text,
    description text,
    owner_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    facile_id text,
    registre_space_id text
);

CREATE SEQUENCE spaces_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE spaces_id_seq OWNED BY spaces.id;

CREATE TABLE users (
    id bigint NOT NULL,
    email text,
    name text,
    password_hash text,
    created_at timestamp with time zone,
    reminder_interval_days bigint DEFAULT 3,
    avatar_url text,
    avatar_source text,
    oidc_picture_url text,
    oidc_access_token text,
    oidc_refresh_token text,
    oidc_token_expiry timestamp with time zone,
    profile_synced_at timestamp with time zone,
    oidc_subject text,
    avatar_upload_path text,
    locale text DEFAULT 'en'::text NOT NULL
);

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

CREATE TABLE webhooks (
    id bigint NOT NULL,
    owner_id bigint,
    url text,
    secret text,
    enabled boolean DEFAULT true,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    space_id bigint
);

CREATE SEQUENCE webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE webhooks_id_seq OWNED BY webhooks.id;

ALTER TABLE ONLY clients ALTER COLUMN id SET DEFAULT nextval('clients_id_seq'::regclass);

ALTER TABLE ONLY documents ALTER COLUMN id SET DEFAULT nextval('documents_id_seq'::regclass);

ALTER TABLE ONLY fields ALTER COLUMN id SET DEFAULT nextval('fields_id_seq'::regclass);

ALTER TABLE ONLY porte_sessions ALTER COLUMN id SET DEFAULT nextval('porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY signers ALTER COLUMN id SET DEFAULT nextval('signers_id_seq'::regclass);

ALTER TABLE ONLY smtp_configs ALTER COLUMN id SET DEFAULT nextval('smtp_configs_id_seq'::regclass);

ALTER TABLE ONLY space_members ALTER COLUMN id SET DEFAULT nextval('space_members_id_seq'::regclass);

ALTER TABLE ONLY spaces ALTER COLUMN id SET DEFAULT nextval('spaces_id_seq'::regclass);

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER TABLE ONLY webhooks ALTER COLUMN id SET DEFAULT nextval('webhooks_id_seq'::regclass);

ALTER TABLE ONLY clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);

ALTER TABLE ONLY directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY fields
    ADD CONSTRAINT fields_pkey PRIMARY KEY (id);

ALTER TABLE ONLY mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY porte_magic_links
    ADD CONSTRAINT porte_magic_links_pkey PRIMARY KEY (token_hash);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY signers
    ADD CONSTRAINT signers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY smtp_configs
    ADD CONSTRAINT smtp_configs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY space_members
    ADD CONSTRAINT space_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY spaces
    ADD CONSTRAINT spaces_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);

CREATE INDEX idx_clients_owner_id ON clients USING btree (owner_id);

CREATE INDEX idx_documents_client_id ON documents USING btree (client_id);

CREATE INDEX idx_documents_original_hash ON documents USING btree (original_hash);

CREATE UNIQUE INDEX idx_documents_owner_external_ref ON documents USING btree (owner_id, external_ref) WHERE ((external_ref)::text <> ''::text);

CREATE INDEX idx_documents_owner_id ON documents USING btree (owner_id);

CREATE INDEX idx_documents_signed_hash ON documents USING btree (signed_hash);

CREATE INDEX idx_documents_space_id ON documents USING btree (space_id);

CREATE INDEX idx_documents_status ON documents USING btree (status);

CREATE INDEX idx_fields_document_id ON fields USING btree (document_id);

CREATE INDEX idx_fields_signer_id ON fields USING btree (signer_id);

CREATE INDEX idx_signers_document_id ON signers USING btree (document_id);

CREATE INDEX idx_signers_token ON signers USING btree (token);

CREATE UNIQUE INDEX idx_smtp_configs_owner_id ON smtp_configs USING btree (owner_id);

CREATE INDEX idx_smtp_configs_space_id ON smtp_configs USING btree (space_id);

CREATE INDEX idx_space_members_space_id ON space_members USING btree (space_id);

CREATE UNIQUE INDEX idx_space_members_space_user ON space_members USING btree (space_id, user_id);

CREATE INDEX idx_space_members_user_id ON space_members USING btree (user_id);

CREATE UNIQUE INDEX idx_spaces_facile_id ON spaces USING btree (facile_id);

CREATE INDEX idx_spaces_owner_id ON spaces USING btree (owner_id);

CREATE UNIQUE INDEX idx_spaces_registre_space_id ON spaces USING btree (registre_space_id);

CREATE UNIQUE INDEX idx_users_email ON users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON users USING btree (oidc_subject);

CREATE INDEX idx_webhooks_owner_id ON webhooks USING btree (owner_id);

CREATE INDEX idx_webhooks_space_id ON webhooks USING btree (space_id);

CREATE INDEX mcp_actions_open_idx ON mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON porte_identities USING btree (user_id);

CREATE INDEX porte_sessions_expiry_idx ON porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON porte_sessions USING btree (user_id);

ALTER TABLE ONLY space_members
    ADD CONSTRAINT fk_space_members_space FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

ALTER TABLE ONLY space_members
    ADD CONSTRAINT fk_space_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_magic_links
    ADD CONSTRAINT porte_magic_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
SELECT 'baseline migration is not reversible' WHERE false;
