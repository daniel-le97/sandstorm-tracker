import type { Action } from '../command';

export const versionAction: Action = () => {
    console.log( 'sandstorm 2.0.0' );
};
