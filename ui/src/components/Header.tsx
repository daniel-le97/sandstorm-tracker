import { useLocation } from 'preact-iso';
import { usePB } from '../hooks/usePB';
import { useState, useEffect } from 'preact/hooks';
import styles from './Header.module.css';

export function Header () {
	const { url } = useLocation();
	const pb = usePB();
	const [ isAuthenticated, setIsAuthenticated ] = useState( false );
	const [ adminEmail, setAdminEmail ] = useState( '' );

	useEffect( () => {
		// Check if admin is authenticated
		const admin = pb.authStore.model;
		if ( admin )
		{
			setIsAuthenticated( true );
			setAdminEmail( admin.email || '' );
		}
	}, [] );

	const handleLogout = () => {
		pb.authStore.clear();
		setIsAuthenticated( false );
		window.location.href = '/login';
	};

	return (
		<header class={ styles.header }>
			<nav class={ styles.nav }>
				<div class={ styles.navLinks }>
					<a href="/" class={ url == '/' ? styles.active : '' }>
						Dashboard
					</a>
					<a href="/servers" class={ url == '/servers' ? styles.active : '' }>
						Servers
					</a>
					<a href="/matches" class={ url == '/matches' ? styles.active : '' }>
						Matches
					</a>
					<a href="/players" class={ url == '/players' ? styles.active : '' }>
						Players
					</a>
					<a href="/weapons" class={ url == '/weapons' ? styles.active : '' }>
						Weapons
					</a>
				</div>

				<div class={ styles.authSection }>
					{ isAuthenticated ? (
						<>
							<span class={ styles.adminEmail }>{ adminEmail }</span>
							<button class={ styles.logoutBtn } onClick={ handleLogout }>
								Logout
							</button>
						</>
					) : (
						<a href="/login" class={ styles.loginBtn }>
							Login
						</a>
					) }
				</div>
			</nav>
		</header>
	);
}
