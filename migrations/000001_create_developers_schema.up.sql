CREATE TABLE developers (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255),
    company_name    VARCHAR(255),
    
    -- Account status
    status          VARCHAR(20) NOT NULL DEFAULT 'active' 
                    CHECK (status IN ('pending', 'active', 'suspended', 'deleted')),
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Subscription info
    plan_tier       VARCHAR(50) DEFAULT 'free',
    
    -- Timestamps
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMP WITH TIME ZONE,
    
    -- Metadata for any additional info
    metadata        JSONB DEFAULT '{}'::jsonb
);