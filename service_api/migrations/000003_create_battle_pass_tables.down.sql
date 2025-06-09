-- Drop indexes
DROP INDEX IF EXISTS idx_battle_passes_status;
DROP INDEX IF EXISTS idx_player_battle_passes_player_id;
DROP INDEX IF EXISTS idx_battle_pass_xp_transactions_player_battle_pass_id;
DROP INDEX IF EXISTS idx_player_battle_pass_rewards_player_battle_pass_id;

-- Drop tables in reverse order of creation (to handle foreign key constraints)
DROP TABLE IF EXISTS battle_pass_xp_transactions;
DROP TABLE IF EXISTS player_battle_pass_rewards;
DROP TABLE IF EXISTS player_battle_passes;
DROP TABLE IF EXISTS battle_pass_levels;
DROP TABLE IF EXISTS battle_passes; 