# Sandstorm Tracker: Game Rules & Logic

## 1. Kill & Assist Logic

- **Kill:**
  - Kills are worth 1 point.
  - The first name in a kill event is the main killer.
  - Killing a teammate is not a kill.
  - Suicides are not kills, but do count as a death.
- **Assist:**
  - Assists are worth 0.5 points.
  - All other players listed in a kill event are assisters.
- **Special Roles:**
  - If a player is a Commander or Observer, their fire support kills are worth 0.5 points.
- **Friendly Fire:**
  - Dying to friendly fire does not increase a player's death counter, but it stops their alive timer.

## 2. Scoring & Ranking

- **Score Calculation:**
  - Player score = (Kills + Assists + Objective Captures/Destroys)
  - For now, scoring is custom; in the future, may use RCON `listplayers`.
- **Ranking:**
  - Player rank is based on score divided by total time alive (not total time played).
- **Commands:**
  - `!kdr`: Get all kills, deaths, and KDR for a player (respond via RCON `say`).
  - `!guns`: Get top 3 weapons with most kills for a player (respond via RCON `say`).
  - `!top`: Get top 3 players on a server (respond via RCON `say`).
  - `!stats`: Get a player's rank and stats (respond via RCON `say`).

## 3. Objective Events

- When players capture or destroy an objective:
  - All dead players are respawned.
  - Alive timers for those players must be restarted.

## 4. Player Status & Server Crashes

If a server crashes:

- All players' alive timers should be **paused** (not stopped), since a crash is not the player's fault.
- Do not count as a death unless a kill event exists for that player.

---

**Summary Table:**

| Event                       | Points | Death Count | Alive Timer | Notes                           |
| --------------------------- | ------ | ----------- | ----------- | ------------------------------- |
| Kill                        | 1      | No          | -           | Main killer only                |
| Assist                      | 0.5    | No          | -           | All other assisters             |
| Fire Support (Cmd/Obs) Kill | 0.5    | No          | -           | Only for Cmd/Obs                |
| Friendly Fire Death         | 0      | No          | Stops       | Not a death, but timer stops    |
| Suicide                     | 0      | Yes         | -           | Not a kill, but is a death      |
| death by bot                | 0      | Yes         | Stops       | regular death                   |
| Teamkill                    | 0      | No          | -           | Not a kill                      |
| Objective Capture/Destroy   | +score | No          | Respawned   | Dead players' timers restart    |
| Server Crash                | 0      | No          | Paused      | Alive timer paused, not a death |
