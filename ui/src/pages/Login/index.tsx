import { usePB } from '../../hooks/usePB';
import { useState } from 'preact/hooks';
import { useLocation } from 'preact-iso';
import styles from './Login.module.css';

export function Login () {
    const pb = usePB();
    const { url } = useLocation();
    const [ email, setEmail ] = useState( '' );
    const [ password, setPassword ] = useState( '' );
    const [ error, setError ] = useState( '' );
    const [ loading, setLoading ] = useState( false );

    const handleLogin = async ( e: Event ) => {
        e.preventDefault();
        setError( '' );
        setLoading( true );

        try
        {
            await pb.collection( '_superusers' ).authWithPassword( email, password );
            // Redirect to dashboard on success
            window.location.href = '/';
        } catch ( err: any )
        {
            setError( err.message || 'Login failed' );
            setLoading( false );
        }
    };

    return (
        <div class={ styles.container }>
            <div class={ styles.loginBox }>
                <h1>Sandstorm Tracker</h1>
                <p>Admin Login</p>

                <form onSubmit={ handleLogin }>
                    { error && <div class={ styles.error }>{ error }</div> }

                    <div class={ styles.formGroup }>
                        <label htmlFor="email">Email</label>
                        <input
                            id="email"
                            type="email"
                            value={ email }
                            onChange={ ( e ) => setEmail( e.currentTarget.value ) }
                            placeholder="admin@example.com"
                            required
                        />
                    </div>

                    <div class={ styles.formGroup }>
                        <label htmlFor="password">Password</label>
                        <input
                            id="password"
                            type="password"
                            value={ password }
                            onChange={ ( e ) => setPassword( e.currentTarget.value ) }
                            placeholder="Enter password"
                            required
                        />
                    </div>

                    <button type="submit" disabled={ loading }>
                        { loading ? 'Logging in...' : 'Login' }
                    </button>
                </form>

                <p class={ styles.info }>
                    For full admin access, use the <a href="/_/">PocketBase Admin Dashboard</a>
                </p>
            </div>
        </div>
    );
}
