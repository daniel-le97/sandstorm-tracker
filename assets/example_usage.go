package assets

// Example usage of the embedded assets:
//
// 1. Access embedded templates (for web routes):
//    assets := assets.GetWebAssets()
//    html, err := registry.LoadFS(assets.FS(), "templates/layout.html", "templates/servers.html")
//
// 2. Get example config content:
//    assets := assets.GetWebAssets()
//    ymlContent, err := assets.GetExampleConfig("yml")
//    tomlContent, err := assets.GetExampleConfig("toml")
//
// 3. Write example config to file:
//    assets := assets.GetWebAssets()
//    err := assets.WriteExampleConfig("./my-config.yml", "yml")
//    err := assets.WriteExampleConfig("./my-config.toml", "toml")
//
// 4. List all embedded configs:
//    assets := assets.GetWebAssets()
//    configs, err := assets.ListConfigs()
//    // Returns: ["sandstorm-tracker.example.yml", "sandstorm-tracker.example.toml"]
//
// 5. From config package (helper functions):
//    import "sandstorm-tracker/internal/app"
//
//    // Check if config exists
//    if !app.ConfigFileExists() {
//        // Generate example config
//        err := app.GenerateExampleConfig("sandstorm-tracker.yml", "yml")
//    }
