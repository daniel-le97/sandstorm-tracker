import { useLocation } from 'preact-iso';
import styles from './Header.module.css';

export function Header () {
	const { url } = useLocation();

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
					<a href="http://localhost:8090/_/" class={ styles.adminDashboardBtn } target="_blank">
						Admin Dashboard
					</a>
				</div>
			</nav>
		</header>
	);
}
