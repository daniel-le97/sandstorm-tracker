#!/usr/bin/env bun

import { existsSync, statSync } from "fs";
import { readFile } from "fs/promises";
import { join } from "path";

interface MapData {
    maps: Array<{
        name: string;
        fileName: string;
        imageUrl: string;
        description: string;
    }>;
    metadata: {
        source: string;
        totalMaps: number;
        lastUpdated: string;
    };
}

async function validateAssets() {
    console.log("🔍 Validating Insurgency: Sandstorm Assets...\n");

    try {
        // Read the maps JSON file
        const mapsData: MapData = JSON.parse(await readFile("./assets/sandstorm-maps.json", "utf-8"));

        console.log(`📊 Found ${mapsData.metadata.totalMaps} maps in database`);
        console.log(`📅 Last updated: ${mapsData.metadata.lastUpdated}`);
        console.log(`🔗 Source: ${mapsData.metadata.source}\n`);

        let validCount = 0;
        let totalSize = 0;

        for (const map of mapsData.maps) {
            const imagePath = join("./assets/maps", map.fileName);
            const exists = existsSync(imagePath);

            if (exists) {
                const stats = statSync(imagePath);
                const sizeKB = Math.round(stats.size / 1024);
                totalSize += stats.size;

                console.log(
                    `✅ ${map.name.padEnd(12)} | ${map.fileName.padEnd(18)} | ${sizeKB.toString().padStart(4)} KB`
                );
                validCount++;
            } else {
                console.log(`❌ ${map.name.padEnd(12)} | ${map.fileName.padEnd(18)} | MISSING`);
            }
        }

        const totalSizeMB = Math.round(totalSize / (1024 * 1024));

        console.log("\n📈 Summary:");
        console.log(`   Valid images: ${validCount}/${mapsData.maps.length}`);
        console.log(`   Total size: ${totalSizeMB} MB`);

        if (validCount === mapsData.maps.length) {
            console.log("\n🎉 All assets validated successfully!");
            return true;
        } else {
            console.log("\n⚠️  Some assets are missing!");
            return false;
        }
    } catch (error) {
        console.error("❌ Error validating assets:", error);
        return false;
    }
}

// Run validation
validateAssets().then((success) => {
    process.exit(success ? 0 : 1);
});
