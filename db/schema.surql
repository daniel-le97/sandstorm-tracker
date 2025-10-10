-- Servers
DEFINE TABLE servers SCHEMAFULL;
DEFINE FIELD external_id ON servers TYPE string ASSERT $value != NONE;
DEFINE FIELD name ON servers TYPE string ASSERT $value != NONE;
DEFINE FIELD path ON servers TYPE option<string>;
DEFINE FIELD created_at ON servers TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON servers TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_servers_external_id ON servers FIELDS external_id UNIQUE;

-- Server logs
DEFINE TABLE server_logs SCHEMAFULL;
DEFINE FIELD server_id ON server_logs TYPE record<servers>;
DEFINE FIELD open_time ON server_logs TYPE datetime ASSERT $value != NONE;
DEFINE FIELD log_path ON server_logs TYPE string ASSERT $value != NONE;
DEFINE FIELD lines_processed ON server_logs TYPE int DEFAULT 0;
DEFINE FIELD file_size_bytes ON server_logs TYPE int DEFAULT 0;
DEFINE FIELD created_at ON server_logs TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON server_logs TYPE datetime DEFAULT time::now();

-- Players
DEFINE TABLE players SCHEMAFULL;
DEFINE FIELD external_id ON players TYPE string ASSERT $value != NONE;
DEFINE FIELD name ON players TYPE string ASSERT $value != NONE;
DEFINE FIELD created_at ON players TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON players TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_players_external_id ON players FIELDS external_id UNIQUE;

-- Matches
DEFINE TABLE matches SCHEMAFULL;
DEFINE FIELD server_id ON matches TYPE record<servers>;
DEFINE FIELD map_id ON matches TYPE option<record<maps>>;
DEFINE FIELD winner_team ON matches TYPE option<int>;
DEFINE FIELD start_time ON matches TYPE option<datetime>;
DEFINE FIELD end_time ON matches TYPE option<datetime>;
DEFINE FIELD mode ON matches TYPE string ASSERT $value != NONE;
DEFINE FIELD created_at ON matches TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON matches TYPE datetime DEFAULT time::now();

-- Match participants
DEFINE TABLE match_participant SCHEMAFULL;
DEFINE FIELD player_id ON match_participant TYPE record<players>;
DEFINE FIELD match_id ON match_participant TYPE record<matches>;
DEFINE FIELD join_time ON match_participant TYPE option<datetime>;
DEFINE FIELD leave_time ON match_participant TYPE option<datetime>;
DEFINE FIELD team ON match_participant TYPE option<int>;
DEFINE FIELD created_at ON match_participant TYPE datetime DEFAULT time::now();

-- Kills
DEFINE TABLE kills SCHEMAFULL;
DEFINE FIELD server_id ON kills TYPE record<servers>;
DEFINE FIELD match_id ON kills TYPE option<record<matches>>;
DEFINE FIELD match_participant_id ON kills TYPE option<record<match_participant>>;
DEFINE FIELD weapon_name ON kills TYPE option<string>;
DEFINE FIELD created_at ON kills TYPE datetime DEFAULT time::now();
DEFINE FIELD killer_id ON kills TYPE option<record<players>>;
DEFINE FIELD victim_name ON kills TYPE option<string>;
DEFINE FIELD kill_type ON kills TYPE int ASSERT $value IN [0,1,2];

-- Player lives
DEFINE TABLE player_lives SCHEMAFULL;
DEFINE FIELD match_participant_id ON player_lives TYPE record<match_participant>;
DEFINE FIELD spawn_time ON player_lives TYPE datetime ASSERT $value != NONE;
DEFINE FIELD death_time ON player_lives TYPE option<datetime>;
DEFINE FIELD cause_of_death ON player_lives TYPE option<string>;

-- Maps
DEFINE TABLE maps SCHEMAFULL;
DEFINE FIELD map_name ON maps TYPE string ASSERT $value != NONE;
DEFINE FIELD scenario ON maps TYPE option<string>;
DEFINE FIELD created_at ON maps TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON maps TYPE datetime DEFAULT time::now();

-- Schema version
DEFINE TABLE schema_version SCHEMAFULL;
DEFINE FIELD version ON schema_version TYPE int ASSERT $value != NONE;
DEFINE FIELD applied_at ON schema_version TYPE datetime DEFAULT time::now();