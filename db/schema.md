servers
- id
- external_id
- name
- path
- created_at
- updated_at

server_logs
 - id
 - server_id
 - open_time
 - log_path
 - lines_processed
 - file_size_bytes
 - created_at
 - updated_at

players
- id
- external_id (steam_id)
- name
- created_at
- updated_at

kills
- id
- server_id
- match_id
- weapon_name
- created_at
- killer_id
- victim_name
- is_team_kill
- is_suicide

maps
 - id
 - map_name
 - scenario
 - created_at
 - updated_at

matches
- id
- server_id
- map_id
- winner_team
- start_time
- end_time
- created_at
- updated_at

match_participant
- id
- player_id
- match_id
- join_time
- leave_time
- team