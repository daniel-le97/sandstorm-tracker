import type { ServerConfig } from "../config";
import { watch } from "fs/promises";


type ServerHandle = {
    config: ServerConfig;
    watcher: AsyncIterableIterator<{ eventType: "change" | "rename"; filename: string | null; }>;
    abortController: AbortController;
    lastProcessedTime: number;
    errorCount: number;
    lastError: Error | null;
    isHealthy: boolean;
    lastHealthCheck: number;
}


class MyService {
    constructor() {
        console.log("MyService initialized");
    }

    async addServerWatcher(config: ServerConfig) {
        const ac = new AbortController();
        const { signal } = ac;
        const dirWatcher = watch(config.logPath, { recursive: false, signal });

        for await (const evt of dirWatcher) {
            if (!evt.filename) continue;
            console.log(`Event on ${evt.filename}: ${evt.eventType}`);

    }
}
}
