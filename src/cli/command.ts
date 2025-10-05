export type FlagType = 'string' | 'boolean';

export type FlagSpec = {
    type: FlagType;
    alias?: string;
    description?: string;
    default?: any;
};

export type FlagsSpec = Record<string, FlagSpec>;

export type ParseResult = {
    flags: Record<string, any>;
    args: string[];
};

export type Action = ( ctx: { flags: Record<string, any>; args: string[]; raw: string[]; } ) => void | Promise<void>;

export class Command {
    name: string;
    description?: string;
    flags: FlagsSpec;
    subcommands: Map<string, Command>;
    action?: Action;

    constructor ( opts: { name: string; description?: string; flags?: FlagsSpec; action?: Action; } ) {
        this.name = opts.name;
        this.description = opts.description;
        this.flags = opts.flags ?? {};
        this.subcommands = new Map();
        this.action = opts.action;
    }

    addCommand ( cmd: Command ) {
        this.subcommands.set( cmd.name, cmd );
        return this;
    }

    command ( name: string, description?: string ) {
        const cmd = new Command( { name, description } );
        this.addCommand( cmd );
        return cmd;
    }

    usage () {
        const lines: string[] = [];
        lines.push( `Usage: ${ this.name } [command] [flags]` );
        if ( this.description ) lines.push( `\n${ this.description }\n` );
        if ( this.subcommands.size )
        {
            lines.push( '\nCommands:' );
            const cmds: Array<{ left: string; right: string; }> = [];
            let maxLeft = 0;
            for ( const cmd of this.subcommands.values() )
            {
                const left = cmd.name;
                const right = cmd.description ?? '';
                cmds.push( { left, right } );
                if ( left.length > maxLeft ) maxLeft = left.length;
            }
            const columnWidth = maxLeft + 2;
            for ( const c of cmds )
            {
                const pad = ' '.repeat( Math.max( 1, columnWidth - c.left.length ) );
                lines.push( `  ${ c.left }${ pad }${ c.right }` );
            }
        }

        const flagKeys = Object.keys( this.flags );
        if ( flagKeys.length )
        {
            lines.push( '\nFlags:' );
            const flagsList: Array<{ left: string; right: string; }> = [];
            let maxFlagLeft = 0;
            for ( const k of flagKeys )
            {
                const f = this.flags[ k ];
                const alias = f.alias ? `, -${ f.alias }` : '';
                const left = `--${ k }${ alias }`;
                let right = f.description ?? '';
                if ( f.default !== undefined )
                {
                    const def = typeof f.default === 'string' ? `"${ f.default }"` : String( f.default );
                    right = right ? `${ right } (default: ${ def })` : `(default: ${ def })`;
                }
                flagsList.push( { left, right } );
                if ( left.length > maxFlagLeft ) maxFlagLeft = left.length;
            }
            const flagColumnWidth = maxFlagLeft + 2;
            for ( const fl of flagsList )
            {
                const pad = ' '.repeat( Math.max( 1, flagColumnWidth - fl.left.length ) );
                // append default value to description if present
                let desc = fl.right;
                // try to preserve empty description
                if ( desc == null ) desc = '';
                // find the flag spec to get default (we don't have direct access here), so embed default into flagsList earlier
                lines.push( `  ${ fl.left }${ pad }${ desc }` );
            }
        }

        return lines.join( '\n' );
    }

    parse ( argv: string[] ): ParseResult {
        const flags: Record<string, any> = {};
        const args: string[] = [];

        // initialize defaults
        for ( const k of Object.keys( this.flags ) )
        {
            flags[ k ] = this.flags[ k ].default;
        }

        let i = 0;
        while ( i < argv.length )
        {
            const token = argv[ i ];
            if ( token === '--' )
            {
                // rest are positional
                args.push( ...argv.slice( i + 1 ) );
                break;
            }

            if ( token.startsWith( '--' ) )
            {
                const [ rawKey, maybeVal ] = token.slice( 2 ).split( '=' );
                const spec = this.flags[ rawKey ];
                if ( !spec )
                {
                    // unknown, treat as arg
                    args.push( token );
                    i++;
                    continue;
                }

                if ( spec.type === 'boolean' )
                {
                    if ( maybeVal === undefined )
                    {
                        flags[ rawKey ] = true;
                        i++;
                        continue;
                    } else
                    {
                        flags[ rawKey ] = maybeVal === 'true';
                        i++;
                        continue;
                    }
                }

                // string
                if ( maybeVal !== undefined )
                {
                    flags[ rawKey ] = maybeVal;
                    i++;
                    continue;
                }

                // take next token as value
                i++;
                if ( i < argv.length )
                {
                    flags[ rawKey ] = argv[ i ];
                    i++;
                    continue;
                } else
                {
                    flags[ rawKey ] = undefined;
                    break;
                }
            }

            if ( token.startsWith( '-' ) && token.length >= 2 )
            {
                // single-dash alias, may be grouped like -abc -> -a -b -c
                const letters = token.slice( 1 ).split( '' );
                let consumedValue = false;
                for ( const l of letters )
                {
                    // find flag with this alias
                    const key = Object.keys( this.flags ).find( k => this.flags[ k ].alias === l );
                    if ( !key )
                    {
                        // unknown, push as arg
                        args.push( `-${ l }` );
                        continue;
                    }
                    const spec = this.flags[ key ];
                    if ( spec.type === 'boolean' )
                    {
                        flags[ key ] = true;
                        continue;
                    }
                    // string flag: take either the rest of the letters or next argv
                    const remaining = letters.slice( letters.indexOf( l ) + 1 ).join( '' );
                    if ( remaining.length )
                    {
                        flags[ key ] = remaining;
                        consumedValue = true;
                        break;
                    } else
                    {
                        // take next token
                        i++;
                        if ( i < argv.length )
                        {
                            flags[ key ] = argv[ i ];
                            consumedValue = true;
                            break;
                        } else
                        {
                            flags[ key ] = undefined;
                        }
                    }
                }
                i++;
                if ( consumedValue ) i++;
                continue;
            }

            // positional or command
            args.push( token );
            i++;
        }

        return { flags, args };
    }

    async run ( argv?: string[] ): Promise<void> {
        const raw = argv ?? ( typeof process !== 'undefined' ? process.argv.slice( 2 ) : [] );

        // if first token is a subcommand, delegate
        if ( raw.length > 0 )
        {
            const first = raw[ 0 ];
            // support built-in `help` subcommand: `help` or `help <subcommand>`
            if ( first === 'help' )
            {
                if ( raw.length === 1 )
                {
                    console.log( this.usage() );
                    return;
                }
                const targetName = raw[ 1 ];
                const target = this.subcommands.get( targetName );
                if ( target )
                {
                    console.log( target.usage() );
                    return;
                }
                console.log( this.usage() );
                return;
            }

            const sub = this.subcommands.get( first );
            if ( sub )
            {
                return sub.run( raw.slice( 1 ) );
            }
        }

        // top-level help flag (only after delegation attempt)
        if ( raw.includes( '--help' ) || raw.includes( '-h' ) )
        {
            console.log( this.usage() );
            return;
        }

        const parsed = this.parse( raw );

        if ( ( parsed.flags[ 'help' ] ?? parsed.flags[ 'h' ] ) === true )
        {
            console.log( this.usage() );
            return;
        }

        if ( !this.action )
        {
            if ( this.subcommands.size )
            {
                console.log( this.usage() );
                return;
            } else
            {
                throw new Error( `no action registered for command ${ this.name }` );
            }
        }

        const ctx = { flags: parsed.flags, args: parsed.args, raw };
        await this.action( ctx );
    }
}
