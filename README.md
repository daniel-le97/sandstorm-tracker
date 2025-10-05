# Sandstorm Multi-Server Tracker

A high-performance multi-server tracker for Insurgency: Sandstorm that monitors multiple servers simultaneously, tracks player statistics, and provides real-time chat commands across Windows and Mac development environments.

## ✨ Features

- **Multi-Server Support**: Monitor up to 6+ Sandstorm servers simultaneously
- **Cross-Platform Development**: Full support for Windows and Mac development environments
- **Real-time Statistics**: Track kills, deaths, playtime, and weapon usage per server
- **Chat Commands**: Players can check their stats with in-game commands
- **Automatic Database Migration**: Seamless upgrade from single-server to multi-server
- **High Performance**: Efficient file watching with debounced event processing
- **Flexible Configuration**: Environment variables or JSON configuration files
- **Development Tools**: Built-in validation, testing, and setup utilities

## �️ Cross-Platform Development Setup

This project is configured to work seamlessly on both Windows and Mac development environments.

### Prerequisites

- **Bun Runtime**: Install from [bun.sh](https://bun.sh)
- **VS Code** (recommended): With our pre-configured workspace settings
- **Git**: For version control

### Quick Setup

Run our automated setup script to configure your development environment:

```bash
# Clone and setup in one command
git clone <repository-url>
cd sandstorm-tracker
bun run setup
```

The setup script will:

1. ✅ Detect your platform (Windows/Mac)
2. ✅ Create required directories
3. ✅ Validate configuration files
4. ✅ Initialize and migrate database
5. ✅ Verify development dependencies
6. ✅ Check VS Code configuration

### Manual Setup

If you prefer manual setup:

```bash
# Install dependencies
bun install

# Validate your environment
bun run validate:config

# Initialize database
bun run migrate:db

# Create example configuration
bun run setup:example
```

## 🚀 Quick Start

### Single Server Setup

For a single server (legacy mode):

```bash
# Windows
set SANDSTORM_LOG_PATH=C:\Program Files\Steam\steamapps\common\sandstorm-server\Insurgency\Saved\Logs
set SANDSTORM_DB_PATH=sandstorm_stats.db

# Mac/Linux
export SANDSTORM_LOG_PATH="/Users/yourname/Library/Application Support/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs"
export SANDSTORM_DB_PATH="sandstorm_stats.db"

# Run the tracker
bun run dev
```

### Multi-Server Setup

1. **Copy example configuration:**

   ```bash
   cp examples/.env.example .env
   ```

2. **Edit paths in `.env` file** to match your server installations:

   **Windows Example:**

   ```bash
   SERVER1_LOG_PATH=C:\Program Files\Steam\steamapps\common\sandstorm-server\Insurgency\Saved\Logs
   SERVER2_LOG_PATH=D:\Games\SandstormServer2\Insurgency\Saved\Logs
   ```

   **Mac Example:**

   ```bash
   SERVER1_LOG_PATH=/Users/yourname/Library/Application Support/Steam/steamapps/common/sandstorm-server/Insurgency/Saved/Logs
   SERVER2_LOG_PATH=/Users/yourname/Games/SandstormServer2/Insurgency/Saved/Logs
   ```

3. **Run the tracker:**
   ```bash
   bun run dev:watch
   ```

## 📁 Platform-Specific Configuration

### Windows Development

- **File Paths**: Use backslashes or forward slashes (auto-normalized)
- **Terminal**: PowerShell or Command Prompt supported
- **VS Code**: Integrated terminal uses PowerShell by default
- **Paths Example**: `C:\Program Files\Steam\...` or `C:/Program Files/Steam/...`

### Mac Development

- **File Paths**: Use forward slashes (Unix-style)
- **Terminal**: Terminal.app or iTerm2 supported
- **VS Code**: Integrated terminal uses zsh/bash by default
- **Paths Example**: `/Users/yourname/Library/Application Support/Steam/...`

### Cross-Platform Features

- **Automatic Path Normalization**: Paths are automatically converted to the correct format
- **Environment Detection**: Platform-specific behavior is handled automatically
- **Consistent Configuration**: `.editorconfig` ensures consistent formatting
- **Universal Scripts**: All npm/bun scripts work identically on both platforms

### 🆔 Automatic Server ID Detection

The system automatically detects server IDs from log file names using the pattern `{server_id}.log`:

- **Log File Naming**: Each server's logs should be named as `{server_id}.log`
  - Example: `main-server.log`, `hardcore-server.log`, `event-server.log`
- **Cross-Platform Support**: Works with any path format (Windows/Unix)
- **Automatic Extraction**: Server ID is extracted from the filename during log processing
- **Validation**: Server IDs are validated to ensure they contain only safe characters

**Examples:**

```bash
# Windows
C:\Servers\Hardcore\Insurgency\Saved\Logs\hardcore-server.log
                                          └── Server ID: "hardcore-server"

# Mac/Linux
/home/user/servers/casual/Insurgency/Saved/Logs/casual-server.log
                                            └── Server ID: "casual-server"
```

**Supported Server ID Formats:**

- Alphanumeric characters: `server1`, `SERVER2`
- Hyphens and underscores: `main-server`, `test_server`
- Mixed case: `CasualServer`, `Hardcore-PvP`
- Length: 1-50 characters

This eliminates the need to manually configure server IDs in most cases - just ensure your log files follow the naming convention!

## 🧪 Development Commands

```bash
# Development with hot reload
bun run dev:watch

# Run tests
bun run test
bun run test:watch

# Build for production
bun run build

# Clean up test databases and logs
bun run clean

# Validate configuration
bun run validate:config

# Database operations
bun run migrate:db
```

## � VS Code Integration

This project includes complete VS Code workspace configuration:

- **Settings**: Platform-specific terminal and formatting preferences
- **Tasks**: Build, test, and run tasks with keyboard shortcuts
- **Launch Configurations**: Debug configurations for different scenarios
- **Extensions**: Recommended extensions for optimal development experience

### Recommended Extensions

Our `.vscode/extensions.json` includes:

- **Bun for Visual Studio Code**: Bun runtime support
- **TypeScript Importer**: Auto-import TypeScript modules
- **Error Lens**: Inline error highlighting
- **Prettier**: Code formatting
- **EditorConfig**: Consistent formatting across editors

## 📊 Configuration Validation

Run configuration validation anytime to check your setup:

```bash
bun run validate:config
```

This will check:

- ✅ Platform compatibility
- ✅ File paths existence and permissions
- ✅ Environment variables
- ✅ Database connectivity
- ✅ VS Code workspace configuration
- ✅ Cross-platform path normalization

## 🗂️ Project Structure

```
sandstorm-tracker/
├── src/
│   ├── cross-platform-utils.ts    # Path and platform utilities
│   ├── validate-config.ts         # Configuration validation
│   ├── migrate-database.ts        # Database migration
│   ├── database.ts               # Database operations
│   ├── events.ts                 # Event parsing
│   └── stats-service.ts          # Statistics service
├── scripts/
│   ├── setup-dev-environment.ts  # Automated setup
│   └── cleanup-test-databases.ts # Test cleanup
├── .vscode/                      # VS Code configuration
├── .editorconfig                 # Cross-platform formatting
├── tests/                        # Test suite
└── examples/                     # Configuration examples
```

## 📖 Configuration Examples

See the [examples directory](./examples/) for detailed configuration options including:

- Multi-server `.env` configuration
- Single-server environment setup
- Docker deployment examples
- Production configuration templates

## 🤝 Contributing

This project welcomes contributions from both Windows and Mac developers:

1. **Setup**: Use `bun run setup` to configure your environment
2. **Code**: Follow the EditorConfig and Prettier formatting
3. **Test**: Run `bun run test` before submitting
4. **Validate**: Use `bun run validate:config` to check your changes

## 📋 Troubleshooting

### Windows Issues

- Ensure PowerShell execution policy allows scripts: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`
- Check Windows Defender exclusions for the project folder
- Use forward slashes in paths for maximum compatibility

### Mac Issues

- Grant file system access if prompted by macOS
- Ensure Xcode Command Line Tools are installed: `xcode-select --install`
- Check file permissions for log directories

### Cross-Platform Issues

- Run `bun run validate:config` to identify platform-specific problems
- Check that all paths use the correct separators for your platform
- Verify that environment variables are set correctly

For more troubleshooting, see our [documentation](./docs/) directory.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

### JSON Configuration

Use `SANDSTORM_CONFIG_PATH` to specify a JSON configuration file. See `examples/six-servers.json` for a complete example with all available options.

## 🎮 Chat Commands

Players can use these commands in-game:

- `!stats` - Show personal statistics for the current server
- `!playtime` - Show playtime for the current server
- `!top` - Show top players on the current server
- `!weapons` - Show weapon usage statistics

## 🗄️ Database Schema

The tracker uses SQLite with automatic migration support:

- **Multi-server aware**: All data includes server identification
- **Backward compatible**: Existing single-server data is preserved
- **Indexed queries**: Optimized for fast lookups
- **Foreign key constraints**: Data integrity enforcement

## 🚀 Performance

Optimized for high-performance multi-server monitoring:

- **Concurrent file watchers**: One per server log directory
- **Debounced event processing**: Configurable debounce timing per server
- **Prepared statements**: All database operations use prepared statements
- **Indexed queries**: Fast lookups with proper database indexing
- **Memory efficient**: Minimal memory footprint per server

## 📊 Migration from Single Server

The tracker automatically migrates existing single-server databases:

1. Detects legacy database format
2. Adds server identification columns
3. Associates existing data with a default server
4. Preserves all historical statistics

## 🤝 Contributing

This project uses [Bun](https://bun.com) - a fast all-in-one JavaScript runtime.

1. Fork the repository
2. Create your feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `bun test`
5. Submit a pull request

## 📝 License

[Add your license information here]
