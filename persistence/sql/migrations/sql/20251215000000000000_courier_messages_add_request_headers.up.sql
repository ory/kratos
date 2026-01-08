ALTER TABLE courier_messages
    ADD COLUMN IF NOT EXISTS request_headers JSONB;