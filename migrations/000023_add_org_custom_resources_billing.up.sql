-- Add per-org custom resource overrides and billing fields.
-- When a custom_* column is NOT NULL, it overrides the plan's value.
-- When NULL, the org inherits the value from its resource plan.
ALTER TABLE organizations
    ADD COLUMN custom_cpu_cores     INTEGER,
    ADD COLUMN custom_ram_mb        INTEGER,
    ADD COLUMN custom_disk_gb       INTEGER,
    ADD COLUMN custom_max_apps      INTEGER,
    ADD COLUMN custom_max_databases INTEGER,
    ADD COLUMN billing_type         VARCHAR(16) NOT NULL DEFAULT 'free'
        CHECK (billing_type IN ('free', 'paid')),
    ADD COLUMN price_monthly_cents  INTEGER,
    ADD COLUMN currency             CHAR(3)     NOT NULL DEFAULT 'USD',
    ADD COLUMN billing_cycle        VARCHAR(16) NOT NULL DEFAULT 'monthly'
        CHECK (billing_cycle IN ('monthly', 'yearly', 'one_time'));

-- Sanity: if billing_type is 'paid' then price must be set (enforced at app layer too).
-- This is left as an application-layer check rather than a CHECK constraint so
-- super admin can flip back to 'free' without clearing the price atomically.
