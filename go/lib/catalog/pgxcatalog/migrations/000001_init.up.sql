-- Create schema
CREATE SCHEMA IF NOT EXISTS cat;

-- Create works table
CREATE TABLE cat.works (
    uuid UUID PRIMARY KEY,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create sources table
CREATE TABLE cat.sources (
    uuid UUID PRIMARY KEY,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create plans table
CREATE TABLE cat.plans (
    uuid UUID PRIMARY KEY,
    kind VARCHAR NOT NULL CHECK (kind <> ''),
    body JSONB NOT NULL CHECK (body <> '{}'::jsonb)
);

-- Create plan_sources table
CREATE TABLE cat.plan_sources (
    plan_uuid UUID NOT NULL REFERENCES cat.plans(uuid) ON DELETE CASCADE,
    source_uuid UUID NOT NULL REFERENCES cat.sources(uuid) ON DELETE CASCADE,
    PRIMARY KEY (plan_uuid, source_uuid)
);

-- Create plan_works table
CREATE TABLE cat.plan_works (
    plan_uuid UUID NOT NULL REFERENCES cat.plans(uuid) ON DELETE CASCADE,
    work_uuid UUID NOT NULL REFERENCES cat.works(uuid) ON DELETE CASCADE,
    PRIMARY KEY (plan_uuid, work_uuid)
);