import type { Action } from '../lib/command';

export const versionAction: Action = () => {
    console.log( 'sandstorm 2.0.0' );
};
