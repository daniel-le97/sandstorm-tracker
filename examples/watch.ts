import { watch } from "fs";

const watcher = watch('./ok', (event, filename) => {
  console.log(`Detected ${event} in ${filename}`);
});

console.log(`Watching for file changes in ${'./ok'}...`);
