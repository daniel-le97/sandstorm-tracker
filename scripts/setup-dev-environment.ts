#!/usr/bin/env bun
/**
 * Cross-platform development environment setup script
 *
 * This script validates the development environment and sets up
 * the database for multi-server tracking across Windows and Mac.
 */

import { execSync } from "child_process";
import { existsSync, mkdirSync } from "fs";
import { join } from "path";
import { PathUtils } from "../src/cross-platform-utils";

interface SetupStep {
    name: string;
    description: string;
    execute: () => Promise<boolean>;
}

class DevEnvironmentSetup {
    private platform = PathUtils.getPlatformInfo().platform;
    private steps: SetupStep[] = [];

    constructor() {
        this.initializeSteps();
    }

    private initializeSteps(): void {
        this.steps = [
            {
                name: "Platform Detection",
                description: "Detect current platform and validate compatibility",
                execute: this.detectPlatform.bind(this),
            },
            {
                name: "Directory Structure",
                description: "Create required directories for cross-platform development",
                execute: this.createDirectories.bind(this),
            },
            {
                name: "Configuration Validation",
                description: "Validate cross-platform configuration files",
                execute: this.validateConfiguration.bind(this),
            },

            {
                name: "Development Dependencies",
                description: "Verify development tools and dependencies",
                execute: this.verifyDependencies.bind(this),
            },
            {
                name: "VS Code Configuration",
                description: "Verify VS Code workspace configuration",
                execute: this.validateVSCodeConfig.bind(this),
            },
        ];
    }

    async run(): Promise<void> {
        console.log("🚀 Setting up cross-platform development environment...");
        console.log(`📱 Platform: ${this.platform}`);
        console.log("");

        let successCount = 0;
        const totalSteps = this.steps.length;

        for (const [index, step] of this.steps.entries()) {
            const stepNumber = index + 1;
            console.log(`[${stepNumber}/${totalSteps}] ${step.name}`);
            console.log(`    ${step.description}`);

            try {
                const success = await step.execute();
                if (success) {
                    console.log(`    ✅ ${step.name} completed successfully`);
                    successCount++;
                } else {
                    console.log(`    ⚠️  ${step.name} completed with warnings`);
                }
            } catch (error) {
                console.error(`    ❌ ${step.name} failed:`, error instanceof Error ? error.message : error);
            }

            console.log("");
        }

        // Final summary
        if (successCount === totalSteps) {
            console.log("🎉 Development environment setup completed successfully!");
            console.log("");
            this.printNextSteps();
        } else {
            console.log(
                `⚠️  Setup completed with ${totalSteps - successCount} issues. Please review the output above.`
            );
        }
    }

    private async detectPlatform(): Promise<boolean> {
        console.log(`    Platform: ${this.platform}`);

        // Check Node.js/Bun version
        try {
            const bunVersion = execSync("bun --version", { encoding: "utf-8" }).trim();
            console.log(`    Bun version: ${bunVersion}`);
        } catch (error) {
            console.warn("    ⚠️  Bun not found. Please install Bun from https://bun.sh");
            return false;
        }

        return true;
    }

    private async createDirectories(): Promise<boolean> {
        const directories = ["logs", "tests/databases", "config", ".vscode"];

        let allCreated = true;

        for (const dir of directories) {
            const fullPath = join(process.cwd(), dir);
            try {
                if (!existsSync(fullPath)) {
                    mkdirSync(fullPath, { recursive: true });
                    console.log(`    Created directory: ${dir}`);
                } else {
                    console.log(`    Directory exists: ${dir}`);
                }
            } catch (error) {
                console.error(`    Failed to create ${dir}:`, error);
                allCreated = false;
            }
        }

        return allCreated;
    }

    private async validateConfiguration(): Promise<boolean> {
        try {
            // Run the configuration validation script
            console.log("    Running configuration validation...");
            execSync("bun run src/validate-config.ts", {
                encoding: "utf-8",
                stdio: "inherit",
            });
            return true;
        } catch (error) {
            console.error("    Configuration validation failed");
            return false;
        }
    }

    private async verifyDependencies(): Promise<boolean> {
        try {
            // Check if node_modules exists
            if (!existsSync("node_modules")) {
                console.log("    Installing dependencies...");
                execSync("bun install", { stdio: "inherit" });
            } else {
                console.log("    Dependencies already installed");
            }

            // Verify TypeScript
            try {
                execSync("bun tsc --noEmit", { encoding: "utf-8" });
                console.log("    TypeScript compilation check passed");
            } catch (error) {
                console.warn("    ⚠️  TypeScript compilation issues detected");
                return false;
            }

            return true;
        } catch (error) {
            console.error("    Dependency verification failed");
            return false;
        }
    }

    private async validateVSCodeConfig(): Promise<boolean> {
        const vscodeFiles = [
            ".vscode/settings.json",
            ".vscode/tasks.json",
            ".vscode/launch.json",
            ".vscode/extensions.json",
        ];

        let allExist = true;

        for (const file of vscodeFiles) {
            if (existsSync(file)) {
                console.log(`    ✓ ${file} exists`);
            } else {
                console.log(`    ⚠️  ${file} missing`);
                allExist = false;
            }
        }

        return allExist;
    }

    private printNextSteps(): void {
        console.log("📋 Next Steps:");
        console.log("");
        console.log("1. 🔧 Configure your Sandstorm servers:");
        console.log("   bun run setup:example");
        console.log("");
        console.log("2. 🚀 Start development:");
        console.log("   bun run dev:watch");
        console.log("");
        console.log("3. 🧪 Run tests:");
        console.log("   bun run test");
        console.log("");
        console.log("4. 📊 Validate configuration anytime:");
        console.log("   bun run validate:config");
        console.log("");

        if (this.platform === "Windows") {
            console.log("💡 Windows-specific tips:");
            console.log("   - Use PowerShell or Command Prompt");
            console.log("   - Ensure paths use backslashes when needed");
            console.log("   - Check Windows Defender exclusions for dev folder");
            console.log("");
        } else if (this.platform === "macOS") {
            console.log("💡 macOS-specific tips:");
            console.log("   - Use Terminal or iTerm2");
            console.log("   - Grant file system access if prompted");
            console.log("   - Consider using Homebrew for additional tools");
            console.log("");
        }
    }
}

// Run setup if this script is executed directly
if (import.meta.main) {
    const setup = new DevEnvironmentSetup();
    await setup.run();
}
