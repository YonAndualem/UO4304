-- Trade License Application schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS applications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_type VARCHAR(50)  NOT NULL,
    applicant_id VARCHAR(255) NOT NULL,
    status       VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    notes        TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_applications_applicant_id ON applications(applicant_id);
CREATE INDEX IF NOT EXISTS idx_applications_status       ON applications(status);

CREATE TABLE IF NOT EXISTS commodities (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           VARCHAR(255) NOT NULL,
    description    TEXT,
    category       VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS documents (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID         NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           VARCHAR(255) NOT NULL,
    url            TEXT         NOT NULL,
    content_type   VARCHAR(100),
    uploaded_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_application_id ON documents(application_id);

CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID           NOT NULL UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
    amount         NUMERIC(15, 2) NOT NULL,
    currency       VARCHAR(10)    NOT NULL DEFAULT 'USD',
    transaction_id VARCHAR(255)   NOT NULL UNIQUE,
    paid_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    status         VARCHAR(50)    NOT NULL DEFAULT 'SETTLED'
);
