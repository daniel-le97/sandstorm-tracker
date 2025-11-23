import { LocationProvider, Router, Route, hydrate, prerender as ssr } from 'preact-iso';

import { Header } from './components/Header';
import { Home } from './pages/Home/index';
import { ServerStatus } from './pages/ServerStatus/index';
import { MatchHistory } from './pages/MatchHistory/index';
import { Players } from './pages/Players/index';
import { Weapons } from './pages/Weapons/index';
import { Login } from './pages/Login/index';
import { NotFound } from './pages/_404';
import { PBProvider } from './hooks/usePB';
import './style.css';

export function App () {
	return (
		<PBProvider>
			<LocationProvider>
				<Header />
				<main>
					<Router>
						<Route path="/" component={ Home } />
						<Route path="/servers" component={ ServerStatus } />
						<Route path="/matches" component={ MatchHistory } />
						<Route path="/players" component={ Players } />
						<Route path="/weapons" component={ Weapons } />
						<Route path="/login" component={ Login } />
						<Route default component={ NotFound } />
					</Router>
				</main>
			</LocationProvider>
		</PBProvider>
	);
}

if ( typeof window !== 'undefined' )
{
	hydrate( <App />, document.getElementById( 'app' ) );
}

export async function prerender ( data ) {
	return await ssr( <App { ...data } /> );
}
