import type { Action } from '../lib/command';
import { initializeApplication, startWatching } from '../../app';

export const serveAction: Action = async ( { flags, args } ) => {
    if ( flags.verbose ) console.log( 'verbose enabled' );
    if ( args.length ) console.log( 'extra args:', args );
    await initializeApplication();
    await startWatching();
};
