import { createContext } from 'preact';
import { useContext, useState, useEffect } from 'preact/hooks';
import pb from '../services/pocketbase';

export const PBContext = createContext( null );

export function usePB () {
    const context = useContext( PBContext );
    if ( !context )
    {
        throw new Error( 'usePB must be used within PBProvider' );
    }
    return context;
}

export function PBProvider ( { children } ) {
    return (
        <PBContext.Provider value={ pb }>
            { children }
        </PBContext.Provider>
    );
}
