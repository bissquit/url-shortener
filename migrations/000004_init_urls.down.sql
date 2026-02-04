DROP index IF EXISTS idx_user_id_deleted;
ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted BOOLEAN;
