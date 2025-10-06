import { appendFile, appendFileSync } from "fs";



// const file = Bun.file('./ok/test.log').writer()
const ok = setInterval( () => {
    appendFileSync('./ok/test.log', 'ok', 'utf-8')
}, 1000 );



process.on( 'SIGINT', async () => {
    clearInterval(ok)
    process.exit( 0 );
} );
process.on( 'SIGTERM', async () => {
    clearInterval(ok)
    process.exit( 0 );
} );

