CREATE TABLE search_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT search_profiles_name_unique UNIQUE (name)
);

CREATE TABLE search_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES search_profiles (id) ON DELETE SET NULL,
    config_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL,
    error_message TEXT,
    law TEXT NOT NULL,
    limit_n INTEGER NOT NULL,
    matched_saved INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CONSTRAINT search_runs_law_check CHECK (law IN ('44', '223', 'both')),
    CONSTRAINT search_runs_matched_saved_nonneg CHECK (matched_saved >= 0),
    CONSTRAINT search_runs_limit_n_positive CHECK (limit_n > 0)
);

CREATE INDEX search_runs_profile_id_idx ON search_runs (profile_id);
CREATE INDEX search_runs_status_idx ON search_runs (status);
CREATE INDEX search_runs_created_at_idx ON search_runs (created_at);

CREATE TABLE purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    law TEXT NOT NULL,
    entity_kind TEXT NOT NULL,
    reg_number TEXT,
    guid TEXT,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    customer_name TEXT NOT NULL DEFAULT '',
    customer_inn TEXT,
    customer_extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    file_links JSONB NOT NULL DEFAULT '[]'::jsonb,
    effective_date TIMESTAMPTZ NOT NULL,
    raw_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT purchases_law_check CHECK (law IN ('44', '223')),
    CONSTRAINT purchases_entity_kind_check CHECK (entity_kind IN ('notice', 'contract'))
);

CREATE UNIQUE INDEX purchases_unique_44_reg_number
    ON purchases (law, entity_kind, reg_number)
    WHERE law = '44' AND reg_number IS NOT NULL;

CREATE UNIQUE INDEX purchases_unique_223_guid
    ON purchases (law, entity_kind, guid)
    WHERE law = '223' AND guid IS NOT NULL;

CREATE INDEX purchases_law_idx ON purchases (law);
CREATE INDEX purchases_effective_date_idx ON purchases (effective_date);
CREATE INDEX purchases_title_idx ON purchases (title);

CREATE TABLE search_run_purchases (
    run_id UUID NOT NULL REFERENCES search_runs (id) ON DELETE CASCADE,
    purchase_id UUID NOT NULL REFERENCES purchases (id) ON DELETE CASCADE,
    attached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, purchase_id)
);

CREATE INDEX search_run_purchases_purchase_id_idx ON search_run_purchases (purchase_id);
CREATE INDEX search_run_purchases_run_attached_at_idx ON search_run_purchases (run_id, attached_at);
