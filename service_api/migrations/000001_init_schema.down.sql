-- Drop tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS player_events;
DROP TABLE IF EXISTS event_schedules;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;