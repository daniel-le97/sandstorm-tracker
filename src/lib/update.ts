/**
 * Helpers to check for updates on GitHub and download release assets.
 *
 * Usage:
 *   import { checkForUpdates, downloadLatestReleaseAsset } from './lib/update';
 *   const info = await checkForUpdates();
 *   if (info.isNew) console.log('Update available:', info.latestVersion);
 */

export type ReleaseAsset = {
    name: string;
    size: number;
    browser_download_url: string;
};

export type UpdateInfo = {
    repo: string;
    latestVersion: string;
    latestTag: string;
    html_url: string;
    body?: string;
    published_at?: string;
    prerelease?: boolean;
    assets: ReleaseAsset[];
    currentVersion?: string;
    isNew: boolean;
};

const DEFAULT_REPO = 'daniel-le97/sandstorm-tracker';

async function fetchLatestRelease ( repo = DEFAULT_REPO ) {
    const url = `https://api.github.com/repos/${ repo }/releases/latest`;
    const res = await fetch( url, {
        headers: {
            Accept: 'application/vnd.github.v3+json',
            'User-Agent': 'sandstorm-tracker-updater',
        },
    } );

    if ( res.status === 404 )
    {
        throw new Error( `No releases found for ${ repo } (404)` );
    }

    if ( !res.ok )
    {
        const txt = await res.text().catch( () => '' );
        throw new Error( `GitHub API error ${ res.status }: ${ txt }` );
    }

    const json = await res.json();
    return json;
}

/**
 * Very small semver comparator. Returns 1 if a>b, -1 if a<b, 0 if equal.
 * Accepts versions like "1.2.3" or "v1.2.3" and ignores build/metadata.
 */
export function compareSemver ( a: string, b: string ): number {
    const norm = ( v: string ) => v.replace( /^v/, '' ).split( /[-+]/ )[ 0 ].split( '.' ).map( n => parseInt( n || '0', 10 ) );
    const A = norm( a );
    const B = norm( b );
    const len = Math.max( A.length, B.length );
    for ( let i = 0; i < len; i++ )
    {
        const ai = A[ i ] || 0;
        const bi = B[ i ] || 0;
        if ( ai > bi ) return 1;
        if ( ai < bi ) return -1;
    }
    return 0;
}

/**
 * Check GitHub Releases for an update. If currentVersion is omitted the function
 * will attempt to read ./package.json and use its `version` field.
 */
export async function checkForUpdates ( currentVersion?: string, repo = DEFAULT_REPO ): Promise<UpdateInfo> {
    if ( !currentVersion )
    {
        try
        {
            const pkgText = await Bun.file( 'package.json' ).text();
            const pkg = JSON.parse( pkgText );
            currentVersion = pkg.version;
        } catch ( error )
        {
            // If package.json can't be read, continue with undefined currentVersion
            currentVersion = undefined;
        }
    }

    const release: any = await fetchLatestRelease( repo );

    const tag = release.tag_name ?? release.name ?? '';
    const latestVersion = tag.replace( /^v/, '' );

    const assets: ReleaseAsset[] = ( release.assets ?? [] ).map( ( a: any ) => ( { name: a.name, size: a.size, browser_download_url: a.browser_download_url } ) );

    const isNew = currentVersion ? compareSemver( latestVersion, currentVersion ) > 0 : true;

    const info: UpdateInfo = {
        repo,
        latestVersion,
        latestTag: tag,
        html_url: release.html_url,
        body: release.body,
        published_at: release.published_at,
        prerelease: release.prerelease,
        assets,
        currentVersion,
        isNew,
    };

    return info;
}

/**
 * Download the first release asset that matches the provided matcher (RegExp or substring).
 * If no asset matches, but the release has a tarball_url or zipball_url the function will
 * download that archive instead.
 *
 * Returns the absolute path to the downloaded file.
 */
export async function downloadLatestReleaseAsset ( matcher?: RegExp | string, outDir?: string, repo = DEFAULT_REPO ): Promise<string> {
    const release: any = await fetchLatestRelease( repo );
    const assets: ReleaseAsset[] = ( release.assets ?? [] ).map( ( a: any ) => ( { name: a.name, size: a.size, browser_download_url: a.browser_download_url } ) );

    let selected: ReleaseAsset | null = null;
    if ( matcher )
    {
        const re = typeof matcher === 'string' ? new RegExp( matcher.replace( /[.*+?^${}()|[\]\\]/g, '\\$&' ) ) : matcher;
        selected = assets.find( a => re.test( a.name ) ) ?? null;
    } else
    {
        selected = assets[ 0 ] ?? null;
    }

    let downloadUrl: string | null = null;
    let filename = '';

    if ( selected )
    {
        downloadUrl = selected.browser_download_url;
        filename = selected.name;
    } else if ( release.tarball_url )
    {
        downloadUrl = release.tarball_url;
        filename = `${ release.tag_name ?? 'release' }.tar.gz`;
    } else if ( release.zipball_url )
    {
        downloadUrl = release.zipball_url;
        filename = `${ release.tag_name ?? 'release' }.zip`;
    }

    if ( !downloadUrl ) throw new Error( 'No downloadable asset or archive found for latest release' );

    const outDirectory = outDir ?? ( process.env.TEMP || process.env.TMP || process.cwd() );
    const outPath = `${ outDirectory }/${ filename }`;

    const res = await fetch( downloadUrl, { headers: { 'User-Agent': 'sandstorm-tracker-updater' } } );
    if ( !res.ok ) throw new Error( `Failed to download asset: ${ res.status }` );

    const ab = await res.arrayBuffer();
    await Bun.file( outPath ).write( new Uint8Array( ab ) );

    return outPath;
}

function preferredExtensionForPlatform () {
    const platform = process.platform; // 'win32', 'darwin', 'linux'
    if ( platform === 'win32' ) return '.zip';
    return '.tar.gz';
}

function findAssetForPlatform ( assets: ReleaseAsset[] ) {
    const ext = preferredExtensionForPlatform();
    // prefer exact extension match
    let found = assets.find( a => a.name.endsWith( ext ) );
    if ( found ) return found;

    // fallback: look for platform keywords
    const platform = process.platform;
    if ( platform === 'win32' )
    {
        found = assets.find( a => /win|windows/i.test( a.name ) );
    } else if ( platform === 'darwin' )
    {
        found = assets.find( a => /mac|darwin|osx/i.test( a.name ) );
    } else
    {
        found = assets.find( a => /linux/i.test( a.name ) );
    }
    return found ?? null;
}

/**
 * Download the platform-appropriate release asset and extract it into `outDir`.
 * Returns the extraction directory.
 */
export async function downloadAndExtractLatestRelease ( outDir?: string, repo = DEFAULT_REPO ): Promise<string> {
    const release: any = await fetchLatestRelease( repo );
    const assets: ReleaseAsset[] = ( release.assets ?? [] ).map( ( a: any ) => ( { name: a.name, size: a.size, browser_download_url: a.browser_download_url } ) );

    const asset = findAssetForPlatform( assets ) ?? assets[ 0 ] ?? null;
    if ( !asset )
    {
        // fallback to tarball/zipball
        if ( release.tarball_url || release.zipball_url )
        {
            const url = release.tarball_url ?? release.zipball_url;
            const defaultName = ( release.tag_name ?? 'release' ) + ( url.endsWith( '.zip' ) ? '.zip' : '.tar.gz' );
            const path = await downloadLatestReleaseAsset( undefined, outDir, repo );
            const dest = outDir ?? ( process.env.TEMP || process.env.TMP || process.cwd() );
            await extractArchive( path, dest );
            return dest;
        }
        throw new Error( 'No assets available to download for this release' );
    }

    const path = await downloadLatestReleaseAsset( asset.name, outDir, repo );
    const dest = outDir ?? ( process.env.TEMP || process.env.TMP || process.cwd() );
    await extractArchive( path, dest );
    return dest;
}

async function extractArchive ( archivePath: string, dest: string ): Promise<void> {
    const isZip = archivePath.endsWith( '.zip' );
    const isTarGz = archivePath.endsWith( '.tar.gz' ) || archivePath.endsWith( '.tgz' );

    if ( isZip )
    {
        if ( process.platform === 'win32' )
        {
            // Use PowerShell Expand-Archive on Windows
            const ps = Bun.spawn( {
                cmd: [ 'powershell', '-NoProfile', '-Command', `Expand-Archive -Force -LiteralPath '${ archivePath }' -DestinationPath '${ dest }'` ],
            } );
            const exitCode = await ps.exited;
            if ( exitCode !== 0 ) throw new Error( 'Failed to extract zip using PowerShell' );
            return;
        } else
        {
            // Use unzip on *nix
            const proc = Bun.spawn( { cmd: [ 'unzip', '-o', archivePath, '-d', dest ] } );
            const exitCode = await proc.exited;
            if ( exitCode !== 0 ) throw new Error( 'Failed to extract zip using unzip' );
            return;
        }
    }

    if ( isTarGz )
    {
        const proc = Bun.spawn( { cmd: [ 'tar', '-xzf', archivePath, '-C', dest ] } );
        const exitCode = await proc.exited;
        if ( exitCode !== 0 ) throw new Error( 'Failed to extract tar.gz using tar' );
        return;
    }

    throw new Error( `Unknown archive format for ${ archivePath }` );
}
