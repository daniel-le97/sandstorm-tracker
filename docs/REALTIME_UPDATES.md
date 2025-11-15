# Real-time Updates with PocketBase JS SDK

This project uses **server-side rendering** for initial page loads, then **PocketBase JS SDK** for real-time updates via WebSockets.

## Architecture

1. **Server Renders HTML** - Go handlers fetch data and render complete HTML pages
2. **Client Subscribes** - JavaScript subscribes to PocketBase collections for updates
3. **Real-time Updates** - Changes automatically sync without polling

## Setup

### 1. PocketBase JS SDK is Self-Hosted

The SDK is embedded in `assets/static/pocketbase.umd.js` and served at `/static/pocketbase.umd.js`.

**No external CDN dependencies!**

### 2. Global PocketBase Client

The `layout.html` template initializes a global `pb` client:

```javascript
const pb = new PocketBase(window.location.origin);
window.pb = pb; // Available globally
```

### 3. Page-Specific Subscriptions

Each page template can define a `{{define "scripts"}}` block for subscriptions:

```html
{{define "scripts"}}
<script>
  document.addEventListener("DOMContentLoaded", function () {
    // Subscribe to collection updates
    pb.collection("matches").subscribe("*", function (e) {
      console.log("Update:", e.action, e.record);

      if (e.action === "update") {
        // Update DOM with new data
        updateMatchDisplay(e.record);
      }
    });

    // Clean up on page unload
    window.addEventListener("beforeunload", function () {
      pb.collection("matches").unsubscribe();
    });
  });
</script>
{{end}}
```

## Collections to Subscribe To

### Server Status / Match Updates

```javascript
pb.collection("matches").subscribe("*", callback);
```

- New matches created
- Round/objective updates
- Match end

### Player Stats (Scores/Kills/Deaths)

```javascript
pb.collection("match_player_stats").subscribe("*", callback);
```

- Score updates every 10 seconds (or on game over)
- Kill/death updates in real-time
- Player joins/leaves

### Weapon Stats

```javascript
pb.collection("match_weapon_stats").subscribe("*", callback);
```

- Weapon usage tracking
- Kill counts per weapon

### Server Records

```javascript
pb.collection("servers").subscribe("*", callback);
```

- Server status changes

## Event Object Structure

When a subscription callback fires, you receive:

```javascript
{
    action: 'create' | 'update' | 'delete',
    record: {
        id: 'record_id',
        // ... all record fields
        created: '2025-11-14 12:00:00',
        updated: '2025-11-14 12:05:00'
    }
}
```

## Example: Update Player Score in Real-Time

### 1. Add data attributes to HTML (in template):

```html
{{range .Players}}
<tr data-player-id="{{.Id}}" data-match-id="{{.MatchId}}">
  <td class="player-name">{{.Name}}</td>
  <td class="player-score">{{.Score}}</td>
  <td class="player-kills">{{.Kills}}</td>
  <td class="player-deaths">{{.Deaths}}</td>
</tr>
{{end}}
```

### 2. Subscribe and update DOM:

```javascript
pb.collection("match_player_stats").subscribe("*", function (e) {
  if (e.action === "update") {
    const row = document.querySelector(
      `[data-player-id="${e.record.player}"][data-match-id="${e.record.match}"]`
    );
    if (row) {
      row.querySelector(".player-score").textContent = e.record.score;
      row.querySelector(".player-kills").textContent = e.record.kills;
      row.querySelector(".player-deaths").textContent = e.record.deaths;

      // Add flash animation
      row.classList.add("updated");
      setTimeout(() => row.classList.remove("updated"), 1000);
    }
  }
});
```

### 3. Add flash animation CSS:

```css
tr.updated {
  background-color: #ff6b3544;
  transition: background-color 0.3s ease;
}
```

## Current Implementation

### Server Status Page (`server_status.html`)

- ✅ Subscribes to `matches` collection
- ✅ Subscribes to `match_player_stats` collection
- ⚠️ Currently reloads page on updates (TODO: update DOM directly)

### Other Pages

- 🔄 Players page - ready for real-time score updates
- 🔄 Matches page - ready for live match status
- 🔄 Weapons page - ready for live weapon stats

## Advantages Over Polling

**Before (30s polling):**

- 30 requests/minute per user
- 30s delay for updates
- Wasted bandwidth for unchanged data

**After (WebSocket subscriptions):**

- 1 WebSocket connection
- Instant updates (< 100ms)
- Only changed data transmitted

## Updating the PocketBase SDK

To update the self-hosted SDK:

```powershell
Invoke-WebRequest -Uri "https://cdn.jsdelivr.net/npm/pocketbase@0.21.5/dist/pocketbase.umd.js" -OutFile "assets/static/pocketbase.umd.js"
```

Then rebuild the app to re-embed the file.

## Testing

1. Start the app: `task dev`
2. Open browser to `http://localhost:8090`
3. Open DevTools Console
4. Watch for subscription logs: `Update: ...`
5. Trigger events (kills, objectives, game over)
6. See real-time updates without refresh

## Troubleshooting

### Subscriptions not working

- Check browser console for errors
- Verify PocketBase SDK loaded: `console.log(window.pb)`
- Check WebSocket connection in Network tab

### Updates delayed

- Normal: Some events debounce for 10 seconds (objectives)
- Check `score_debouncer.go` for timing logic

### Page reloads instead of updating

- This is current behavior (safe fallback)
- TODO: Implement DOM updates for each page

## Next Steps

1. ✅ Add data attributes to all table rows (`data-id`, etc.)
2. ✅ Implement DOM update functions instead of page reload
3. ✅ Add visual feedback for updates (flash animations)
4. ✅ Subscribe on all pages (players, weapons, matches)
5. ✅ Add loading states for initial render
