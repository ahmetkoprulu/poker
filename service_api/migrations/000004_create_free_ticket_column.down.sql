ALTER TABLE player_events DROP COLUMN free_tickets;
ALTER TABLE player_events DROP COLUMN event_id;
ALTER TABLE events DROP COLUMN general_config;
ALTER TABLE players DROP COLUMN spins;
ALTER TABLE players DROP COLUMN gold_spins;
ALTER TABLE player_events DROP CONSTRAINT fk_player_events_event_id;