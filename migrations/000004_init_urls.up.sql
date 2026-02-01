ALTER TABLE urls ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_id_deleted ON urls(user_id, is_deleted);
