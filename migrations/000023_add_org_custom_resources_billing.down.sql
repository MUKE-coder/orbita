ALTER TABLE organizations
    DROP COLUMN IF EXISTS custom_cpu_cores,
    DROP COLUMN IF EXISTS custom_ram_mb,
    DROP COLUMN IF EXISTS custom_disk_gb,
    DROP COLUMN IF EXISTS custom_max_apps,
    DROP COLUMN IF EXISTS custom_max_databases,
    DROP COLUMN IF EXISTS billing_type,
    DROP COLUMN IF EXISTS price_monthly_cents,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS billing_cycle;
