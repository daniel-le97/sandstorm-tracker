import PocketBase from 'pocketbase';

// Initialize PocketBase client
// In dev mode (localhost:5173), connect to localhost:8090
// In production (embedded), use the same origin
const getPBUrl = () => {
    if ( typeof window === 'undefined' )
    {
        return 'http://127.0.0.1:8090';
    }

    // Dev server connects to PocketBase on 8090
    if ( window.location.hostname === 'localhost' && window.location.port === '5173' )
    {
        return 'http://localhost:8090';
    }

    // Production/embedded uses same origin
    return window.location.origin;
};

const pb = new PocketBase( getPBUrl() );

export default pb;
