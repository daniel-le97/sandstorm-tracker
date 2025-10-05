# Valibot

The modular and type safe schema library for validating structural data.

## Get started (guides)

### Introduction

Hello, I am Valibot and I would like to help you validate data easily using a schema. No matter if it is incoming data on a server, a form or even configuration files. I have no dependencies and can run in any JavaScript environment.

> I highly recommend you read the [announcement post](https://www.builder.io/blog/introducing-valibot), and if you are a nerd like me, the [bachelor's thesis](/thesis.pdf) I am based on.

#### Highlights

- Fully type safe with static type inference
- Small bundle size starting at less than 700 bytes
- Validate everything from strings to complex objects
- Open source and fully tested with 100 % coverage
- Many transformation and validation actions included
- Well structured source code without dependencies
- Minimal, readable and well thought out API

#### Example

First you create a schema that describes a structured data set. A schema can be compared to a type definition in TypeScript. The big difference is that TypeScript types are "not executed" and are more or less a DX feature. A schema on the other hand, apart from the inferred type definition, can also be executed at runtime to guarantee type safety of unknown data.

{/* prettier-ignore */}
```ts
import * as v from 'valibot'; // 1.31 kB

// Create login schema with email and password
const LoginSchema = v.object({
  email: v.pipe(v.string(), v.email()),
  password: v.pipe(v.string(), v.minLength(8)),
});

// Infer output TypeScript type of login schema as
// { email: string; password: string }
type LoginData = v.InferOutput<typeof LoginSchema>;

// Throws error for email and password
const output1 = v.parse(LoginSchema, { email: '', password: '' });

// Returns data as { email: string; password: string }
const output2 = v.parse(LoginSchema, {
  email: 'jane@example.com',
  password: '12345678',
});
```

Apart from <Link href="/api/parse/">`parse`</Link> I also offer a non-exception-based API with <Link href="/api/safeParse/">`safeParse`</Link> and a type guard function with <Link href="/api/is/">`is`</Link>. You can read more about it <Link href="/guides/parse-data/">here</Link>.

#### Comparison

Instead of relying on a few large functions with many methods, my API design and source code is based on many small and independent functions, each with just a single task. This modular design has several advantages.

For example, this allows a bundler to use the import statements to remove code that is not needed. This way, only the code that is actually used gets into your production build. This can reduce the bundle size by up to 95 % compared to [Zod](https://zod.dev/).

In addition, it allows you to easily extend my functionality with external code and makes my source code more robust and secure because the functionality of the individual functions can be tested much more easily through unit tests.

#### Credits

My friend [Fabian](https://github.com/fabian-hiller) created me as part of his bachelor thesis at [Stuttgart Media University](https://www.hdm-stuttgart.de/en/), supervised by Walter Kriha, [Miško Hevery](https://github.com/mhevery) and [Ryan Carniato](https://github.com/ryansolid). My role models also include [Colin McDonnell](https://github.com/colinhacks), who had a big influence on my API design with [Zod](https://zod.dev/).

#### Feedback

Find a bug or have an idea how to improve my code? Please fill out an [issue](https://github.com/fabian-hiller/valibot/issues/new). Together we can make the library even better!

#### License

I am completely free and licensed under the [MIT license](https://github.com/fabian-hiller/valibot/blob/main/LICENSE.md). But if you like, you can feed me with a star on [GitHub](https://github.com/fabian-hiller/valibot).

### Installation

Valibot is currently available for Node, Bun and Deno. Below you will learn how to add the library to your project.

#### General

Except for this guide, the rest of this documentation assumes that you are using npm for the import statements in the code examples.

It should make no difference whether you use individual imports or a wildcard import. Tree shaking and code splitting should work in both cases.

If you are using TypeScript, we recommend that you enable strict mode in your `tsconfig.json` so that all types are calculated correctly.

> The minimum required TypeScript version is v5.0.2.

```js
{
  "compilerOptions": {
    "strict": true,
    // ...
  }
}
```

#### From npm

For Node and Bun, you can add the library to your project with a single command using your favorite package manager.

```bash
npm install valibot     # npm
yarn add valibot        # yarn
pnpm add valibot        # pnpm
bun add valibot         # bun
```

Then you can import it into any JavaScript or TypeScript file.

```ts
// With individual imports
import { … } from 'valibot';

// With a wildcard import
import * as v from 'valibot';
```

#### From JSR

For Node, Deno and Bun, you can add the library to your project with a single command using your favorite package manager.

```bash
deno add jsr:@valibot/valibot      # deno
npx jsr add @valibot/valibot       # npm
yarn dlx jsr add @valibot/valibot  # yarn
pnpm dlx jsr add @valibot/valibot  # pnpm
bunx jsr add @valibot/valibot      # bun
```

Then you can import it into any JavaScript or TypeScript file.

```ts
// With individual imports
import { … } from '@valibot/valibot';

// With a wildcard import
import * as v from '@valibot/valibot';
```

In Deno, you can also directly reference me using `jsr:` specifiers.

```ts
// With individual imports
import { … } from 'jsr:@valibot/valibot';

// With a wildcard import
import * as v from 'jsr:@valibot/valibot';
```

#### From Deno

With Deno, you can reference the library directly through our deno.land/x URL.

```ts
// With individual imports
import { … } from 'https://deno.land/x/valibot/mod.ts';

// With a wildcard import
import * as v from 'https://deno.land/x/valibot/mod.ts';
```

### Quick start

A Valibot schema can be compared to a type definition in TypeScript. The big difference is that TypeScript types are "not executed" and are more or less a DX feature. A schema on the other hand, apart from the inferred type definition, can also be executed at runtime to truly guarantee type safety of unknown data.

#### Basic concept

Similar to how types can be defined in TypeScript, Valibot allows you to define a schema with various small functions. This applies to primitive values like strings as well as more complex data sets like objects.

```ts
import * as v from 'valibot';

// TypeScript
type LoginData = {
  email: string;
  password: string;
};

// Valibot
const LoginSchema = v.object({
  email: v.string(),
  password: v.string(),
});
```

#### Pipelines

In addition, pipelines enable you to perform more detailed validations and transformations with the <Link href="/api/pipe/">`pipe`</Link> method. Thus, for example, it can be ensured that a string is an email that ends with a certain domain.

```ts
import * as v from 'valibot';

const EmailSchema = v.pipe(v.string(), v.email(), v.endsWith('@example.com'));
```

A pipeline must always start with a schema, followed by up to 19 validation or transformation actions. They are executed in sequence, and the result of the previous action is passed to the next. More details about pipelines can be found in <Link href="/guides/pipelines/">this guide</Link>.

#### Error messages

If an issue is detected during validation, the library emits a specific issue object that includes various details and an error message. This error message can be overridden via the first optional argument of a schema or validation action.

```ts
import * as v from 'valibot';

const LoginSchema = v.object({
  email: v.pipe(
    v.string('Your email must be a string.'),
    v.nonEmpty('Please enter your email.'),
    v.email('The email address is badly formatted.')
  ),
  password: v.pipe(
    v.string('Your password must be a string.'),
    v.nonEmpty('Please enter your password.'),
    v.minLength(8, 'Your password must have 8 characters or more.')
  ),
});
```

Custom error messages allow you to improve the usability of your software by providing specific troubleshooting information and returning error messages in a language other than English. See the <Link href="/guides/internationalization/">i18n guide</Link> for more information.

#### Usage

Finally, you can use your schema to infer its input and output types and to parse unknown data. This way, your schema is the single source of truth. This concept simplifies your development process and makes your code more robust in the long run.

```ts
import * as v from 'valibot';

const LoginSchema = v.object({…});

type LoginData = v.InferOutput<typeof LoginSchema>;

function getLoginData(data: unknown): LoginData {
  return v.parse(LoginSchema, data);
}
```

### Use cases

Next, we would like to point out some use cases for which Valibot is particularly well suited. We welcome [ideas](https://github.com/fabian-hiller/valibot/issues/new) for other use cases that we may not have thought of yet.

#### Server requests

Since most API endpoints can be reached via the Internet, basically anyone can send a request and transmit data. It is therefore important to apply zero trust security and to check request data thoroughly before processing it further.

This works particularly well with a schema, compared to if/else conditions, as even complex structures can be easily mapped. In addition, the library automatically type the parsed data according to the schema, which improves type safety and thus makes your code more secure.

#### Form validation

A schema can also be used for form validation. Due to Valibot's small bundle size and the possibility to individualize the error messages, the library is particularly well suited for this. Also, fullstack frameworks like Next.js, Remix, and Nuxt allow the same schema to be used for validation in the browser as well as on the server, which reduces your code to the minimum.

[Modular Forms](https://modularforms.dev/react/guides/validate-your-fields#schema-validation), for example, offers validation based on a schema at form and field level. In addition, the form can be made type-safe using the schema, which also enables autocompletion during development. In combination with the right framework, a fully type-safe and progressively enhanced form can be created with few lines of code and a great experience for developers and end-users.

#### Browser state

The browser state, which is stored using cookies, search parameters or the local storage, can be accidentally or intentionally manipulated by the user. To ensure the functionality of an application, it can help to validate this data before processing. Valibot can be used for this, which also improves type safety.

#### Config files

Library authors can also make use of Valibot, for example, to match configuration files with a schema and, in the event of an error, provide clear indications of the cause and how to fix the problem. The same applies to environment variables to quickly detect configuration errors.

#### Schema builder

Our schemas are plain JavaScript objects with a well-defined and fully type-safe structure. This makes Valibot a great choice for defining data structures that can be further processed by third-party code. For example, it is possible to build an ORM with custom metadata actions on top of Valibot to generate database schemas. Another example is our official `toJsonSchema` function, which uses Valibot's object API to output a JSON Schema that can be used for documentation purposes or to generate structured output with LLMs.

#### Data migration

Valibot can also be used to migrate data from one form to another in a type-safe way. The advantage of a schema library like Valibot is that transformations can be defined for individual properties instead of for the entire dataset. This can make data migrations more readable and maintainable. In addition, the schema can be used to validate the data before the migration, which increases the reliability of the migration process.

### Comparison

Even though Valibot's API resembles other solutions at first glance, the implementation and structure of the source code is very different. In the following, we would like to highlight the differences that can be beneficial for both you and your users.

#### Modular design

Instead of relying on a few large functions with many methods, Valibot's API design and source code is based on many small and independent functions, each with just a single task. This modular design has several advantages.

On one hand, the functionality of Valibot can be easily extended with external code. On the other, it makes the source code more robust and secure because the functionality of the individual functions as well as special edge cases can be tested much easier through unit tests.

However, perhaps the biggest advantage is that a bundler can use the static import statements to remove any code that is not needed. Thus, only the code that is actually used ends up in the production build. This allows us to extend the functionality of the library with additional functions without increasing the bundle size for all users.

This can make a big difference, especially for client-side validation, as it reduces the bundle size and, depending on the framework, speeds up the startup time.

{/* prettier-ignore */}
```ts
import * as v from 'valibot'; // 1.37 kB

const LoginSchema = v.object({
  email: v.pipe(
    v.string(),
    v.nonEmpty('Please enter your email.'),
    v.email('The email address is badly formatted.')
  ),
  password: v.pipe(
    v.string(),
    v.nonEmpty('Please enter your password.'),
    v.minLength(8, 'Your password must have 8 characters or more.')
  ),
});
```

##### Comparison with Zod

For example, to validate a simple login form, [Zod](https://zod.dev/) requires [13.5 kB](https://bundlejs.com/?q=zod&treeshake=%5B%7B+object%2Cstring+%7D%5D) whereas Valibot require only [1.37 kB](https://bundlejs.com/?q=valibot&treeshake=%5B%7B+email%2CminLength%2CnonEmpty%2Cobject%2Cstring%2Cpipe+%7D%5D). That's a 90 % reduction in bundle size. This is due to the fact that Zod's functions have several methods with additional functionalities, that cannot be easily removed by current bundlers when they are not executed in your source code.

{/* prettier-ignore */}
```ts
import { object, string } from 'zod'; // 13.5 kB

const LoginSchema = object({
  email: string()
    .min(1, 'Please enter your email.')
    .email('The email address is badly formatted.'),
  password: string()
    .min(1, 'Please enter your password.')
    .min(8, 'Your password must have 8 characters or more.'),
});
```

> You can migrate from Zod to Valibot using our [migration guide](/guides/migrate-from-zod/). It provides a codemod and a detailed overview of the differences between the two libraries.

#### Performance

With a schema library, a distinction must be made between startup performance and runtime performance. Startup performance describes the time required to load and initialize the library. This benchmark is mainly influenced by the bundle size and the amount of work required to create a schema. Runtime performance describes the time required to validate unknown data using a schema.

Since Valibot's implementation is optimized to minimize the bundle size and the effort of initialization, there is hardly any library that performs better in a [TTI](https://web.dev/articles/tti) benchmark. In terms of runtime performance, Valibot is in the midfield. Roughly speaking, the library is about twice as fast as [Zod](https://zod.dev/), but much slower than [Typia](https://typia.io/) and [TypeBox](https://github.com/sinclairzx81/typebox), because we don't yet use a compiler that can generate highly optimized runtime code, and our implementation doesn't allow the use of the [`Function`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function/Function) constructor.

> Further details on performance can be found in the [bachelor's thesis](/thesis.pdf) Valibot is based on.

### Ecosystem

This page is for you if you are looking for frameworks or libraries that support Valibot.

> Use the button at the bottom left of this page to add your project to this ecosystem page. Please make sure to add your project to an appropriate existing category in alphabetical order or create a new category if necessary.

#### Frameworks

- [NestJS](https://docs.nestjs.com): A progressive Node.js framework for building efficient, reliable and scalable server-side applications
- [Qwik](https://qwik.dev): A web framework which helps you build instantly-interactive web apps at any scale without effort.

#### API libraries

- [Drizzle ORM](https://orm.drizzle.team/): TypeScript ORM that feels like writing SQL
- [GQLoom](https://gqloom.dev/): Weave GraphQL schema and resolvers using Valibot
- [Hono](https://hono.dev/): Ultrafast web framework for the Edges
- [next-safe-action](https://next-safe-action.dev) Type safe and validated Server Actions for Next.js
- [oRPC](https://orpc.unnoq.com/): Typesafe APIs Made Simple
- [piying-orm](https://github.com/piying-org/piying-orm): ORM for Valibot; Supports TypeORM, with more to come.
- [tRPC](https://trpc.io/): Move Fast and Break Nothing. End-to-end typesafe APIs made easy
- [upfetch](https://github.com/L-Blondy/up-fetch): Advanced fetch client builder

#### AI libraries

- [AI SDK](https://sdk.vercel.ai/): Build AI-powered applications with React, Svelte, Vue, and Solid

#### Form libraries

- [@rvf/valibot](https://github.com/airjp73/rvf/tree/main/packages/valibot): Valibot schema parser for [RVF](https://rvf-js.io/)
- [conform](https://conform.guide/): A type-safe form validation library utilizing web fundamentals to progressively enhance HTML Forms with full support for server frameworks like Remix and Next.js.
- [mantine-form-valibot-resolver](https://github.com/Songkeys/mantine-form-valibot-resolver): Valibot schema resolver for [@mantine/form](https://mantine.dev/form/use-form/)
- [maz-ui](https://maz-ui.com/composables/use-form-validator): Vue3 flexible and typed composable to manage forms simply with multiple modes and advanced features
- [Modular Forms](https://modularforms.dev/): Modular and type-safe form library for SolidJS, Qwik, Preact and React
- [piying-view](https://github.com/piying-org/piying-view): Frontend Form Solution; Supports Angular, Vue, React, with more to come.
- [React Hook Form](https://react-hook-form.com/): React Hooks for form state management and validation
- [regle](https://github.com/victorgarciaesgi/regle): Headless form validation library for Vue.js
- [Superforms](https://superforms.rocks): A comprehensive SvelteKit form library for server and client validation
- [svelte-jsonschema-form](https://x0k.dev/svelte-jsonschema-form/validators/valibot/): Svelte 5 library for creating forms based on JSON schema
- [TanStack Form](https://tanstack.com/form): Powerful and type-safe form state management for the web
- [VeeValidate](https://vee-validate.logaretm.com/v4/): Painless Vue.js forms
- [vue-valibot-form](https://github.com/IlyaSemenov/vue-valibot-form): Minimalistic Vue3 composable for handling form submit

#### Component libraries

- [Nuxt UI](https://ui.nuxt.com/): Fully styled and customizable components for Nuxt

#### Valibot to X

- [@gcornut/cli-valibot-to-json-schema](https://github.com/gcornut/cli-valibot-to-json-schema): CLI wrapper for @valibot/to-json-schema
- [@valibot/to-json-schema](https://github.com/fabian-hiller/valibot/tree/main/packages/to-json-schema): The official JSON schema converter for Valibot
- [Hono OpenAPI](https://github.com/rhinobase/hono-openapi): A plugin for Hono to generate OpenAPI Swagger documentation
- [TypeMap](https://github.com/sinclairzx81/typemap/): Uniform Syntax, Mapping and Compiler Library for TypeBox, Valibot and Zod
- [TypeSchema](https://typeschema.com/): Universal adapter for schema validation
- [Valibot-Fast-Check](https://github.com/Eronmmer/valibot-fast-check): A library to generate [fast-check](https://fast-check.dev) arbitraries from Valibot schemas for property-based testing
- [valibot-serialize](https://github.com/gadicc/valibot-serialize): Serialize a schema to JSON and back again, or to (tree-shaking safe) static code

#### X to Valibot

- [@hey-api/openapi-ts](https://heyapi.dev/openapi-ts/plugins/valibot): The OpenAPI to TypeScript codegen. Generate clients, SDKs, validators, and more.
- [@traversable/valibot](https://github.com/traversable/schema/tree/main/packages/valibot): Build your own "Valibot to X" library, or pick one of 10+ off-the-shelf transformers
- [DRZL](https://github.com/use-drzl/drzl): Analyze Drizzle ORM schema(s) and auto-generate Valibot validators, typed services, and strongly typed routers (oRPC/tRPC/etc) via a modular pipeline.
- [graphql-codegen-typescript-validation-schema](https://github.com/Code-Hex/graphql-codegen-typescript-validation-schema): GraphQL Code Generator plugin to generate form validation schema from your GraphQL schema.
- [Prisma Valibot Generator](https://github.com/omar-dulaimi/prisma-valibot-generator): Generate Valibot validators from your Prisma schema so types and runtime stay in sync.
- [TypeBox-Codegen](https://sinclairzx81.github.io/typebox-workbench/): Code generation for schema libraries
- [TypeMap](https://github.com/sinclairzx81/typemap/): Uniform Syntax, Mapping and Compiler Library for TypeBox, Valibot and Zod
- [valibot-serialize](https://github.com/gadicc/valibot-serialize): From serialized JSON back to a schema instance or the (tree-shaking safe) code to create that instance

#### Utilities

- [@camflan/valibot-openapi-generator](https://github.com/camflan/valibot-openapi-generator): Functions to help build OpenAPI documentation using Valibot schemas
- [@nest-lab/typeschema](https://github.com/jmcdo29/nest-lab/tree/main/packages/typeschema): A ValidationPipe that handles many schema validators in a class-based fashion for NestJS's input validation
- [@traversable/valibot-test](https://github.com/traversable/schema/tree/main/packages/valibot-test): Random Valibot schema generator built for fuzz testing, includes generators for both valid and invalid data
- [@valibot/i18n](https://github.com/fabian-hiller/valibot/tree/main/packages/i18n): The official i18n translations for Valibot
- [fastify-type-provider-valibot](https://github.com/qlaffont/fastify-type-provider-valibot): Fastify Type Provider with Valibot
- [valibot-env](https://y-hiraoka.github.io/valibot-env): Environment variables validator with Valibot
- [valibotx](https://github.com/IlyaSemenov/valibotx): A collection of extensions and shortcuts to core Valibot functions
- [valiload](https://github.com/JuerGenie/valiload): A simple and lightweight library for overloading functions in TypeScript
- [valimock](https://github.com/saeris/valimock): Generate mock data using your Valibot schemas using [Faker](https://github.com/faker-js/faker)
- [valipass](https://github.com/Saeris/valipass): Collection of password validation actions for Valibot schemas

### LLMs.txt

If you are using AI to generate Valibot schemas, you can use our LLMs.txt files to help the AI better understand the library.

#### What is LLMs.txt?

An [LLMs.txt](https://llmstxt.org/) file is a plain text file that provides instructions or metadata for large language models (LLMs). It often specifies how the LLMs should process or interact with content. It is similar to a robots.txt file, but is tailored for AI models.

#### Available routes

We provide several LLMs.txt routes. Use the route that works best with your AI tool.

- [`llms.txt`](/llms.txt) contains a table of contents with links to Markdown files
- [`llms-full.txt`](/llms-full.txt) contains the Markdown content of the entire docs
- [`llms-guides.txt`](/llms-guides.txt) contains the Markdown content of the guides
- [`llms-api.txt`](/llms-api.txt) contains the Markdown content of the API reference

#### How to use it

To help you get started, here are some examples of how the LLMs.txt files can be used with various AI tools.

> Please help us by adding more examples of other AI tools. If you use a tool that supports LLMs.txt files, please [open a pull request](https://github.com/fabian-hiller/valibot/pulls) to add it to this page.

##### Cursor

You can add a custom documentation as context in Cursor using the `@Docs` feature. Read more about it in the [here](https://docs.cursor.com/context/@-symbols/@-docs).

## Main concepts (guides)

### Mental model

Valibot's mental model is mainly divided between **schemas**, **methods**, and **actions**. Since each functionality is imported as its own function, it is crucial to understand this concept as it makes working with the modular API design much easier.

<MentalModelDark
  alt="Code example with a schema, method and actions"
  class="hidden dark:block"
/>
<MentalModelLight
  alt="Code example with a schema, method and actions"
  class="dark:hidden"
/>

> The <Link href="/api/">API reference</Link> gives you a great overview of all schemas, methods, and actions. For each one, the corresponding reference page also lists down other related schemas, methods, and actions for better discoverability.

#### Schemas

Schemas are the starting point for using Valibot. They allow you to validate **a specific data type**, like a string, object, or date. Each schema is independent. They can be reused or even nested to reflect more complex data structures.

```ts
import * as v from 'valibot';

const BookSchema = v.object({
  title: v.string(),
  numberOfPages: v.number(),
  publication: v.date(),
  tags: v.array(v.string()),
});
```

Every schema function returns an accesible object that contains all its properties. However, in most cases you don't need to access them directly. Instead, you use methods that help you modify or use a schema.

#### Methods

Methods help you either **modify or use a schema**. For example, the <Link href="/api/parse/">`parse`</Link> method helps you parse unknown data based on a schema. When you use a method, you always pass the schema as the first argument.

```ts
import * as v from 'valibot';

const BookSchema = v.object({…});

function createBook(data: unknown) {
  return v.parse(BookSchema, data);
}
```

> Most methods are used with schemas. However, there are a few exceptions, such as <Link href="/api/forward/">`forward`</Link> and <Link href="/api/flatten/">`flatten`</Link>, which are used with actions or issues.

#### Actions

Actions help you to **further validate or transform** a specific data type. They are used exclusively in conjunction with the <Link href="/api/pipe/">`pipe`</Link> method, which extends the functionality of a schema by adding additional validation and transformation rules. For example, the following schema can be used to trim a string and check if it is a valid email address.

```ts
import * as v from 'valibot';

const EmailSchema = v.pipe(v.string(), v.trim(), v.email());
```

Actions are very powerful. There are basically no limits to what you can do with them. Besides basic validations and transformations as shown in the example above, they also allow you to modify the output type with actions like <Link href="/api/readonly/">`readonly`</Link> and <Link href="/api/brand/">`brand`</Link>.

### Schemas

Schemas allow you to validate a specific data type. They are similar to type definitions in TypeScript. Besides primitive values like strings and complex values like objects, Valibot also supports special cases like literals, unions and custom types.

#### Primitive values

Valibot supports the creation of schemas for any primitive data type. These are immutable values that are stored directly in the stack, unlike objects where only a reference to the heap is stored.

<ApiList
  label="Primitive schemas"
  items={[
    'bigint',
    'boolean',
    'null',
    'number',
    'string',
    'symbol',
    'undefined',
  ]}
/>

```ts
import * as v from 'valibot';

const BigintSchema = v.bigint(); // bigint
const BooleanSchema = v.boolean(); // boolean
const NullSchema = v.null(); // null
const NumberSchema = v.number(); // number
const StringSchema = v.string(); // string
const SymbolSchema = v.symbol(); // symbol
const UndefinedSchema = v.undefined(); // undefined
```

#### Complex values

Among complex values, Valibot supports objects, records, arrays, tuples, and several other classes.

> There are various methods for objects such as <Link href="/api/pick/">`pick`</Link>, <Link href="/api/omit/">`omit`</Link>, <Link href="/api/partial/">`partial`</Link> and <Link href="/api/required/">`required`</Link>. Learn more about them <Link href="/guides/methods/#object-methods">here</Link>.

<ApiList
  label="Complex schemas"
  items={[
    'array',
    'blob',
    'date',
    'file',
    'function',
    'looseObject',
    'looseTuple',
    'map',
    'object',
    'objectWithRest',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
  ]}
/>

```ts
import * as v from 'valibot';

const ArraySchema = v.array(v.string()); // string[]
const BlobSchema = v.blob(); // Blob
const DateSchema = v.date(); // Date
const FileSchema = v.file(); // File
const FunctionSchema = v.function(); // (...args: unknown[]) => unknown
const LooseObjectSchema = v.looseObject({ key: v.string() }); // { key: string }
const LooseTupleSchema = v.looseTuple([v.string(), v.number()]); // [string, number]
const MapSchema = v.map(v.string(), v.number()); // Map<string, number>
const ObjectSchema = v.object({ key: v.string() }); // { key: string }
const ObjectWithRestSchema = v.objectWithRest({ key: v.string() }, v.null()); // { key: string } & { [key: string]: null }
const PromiseSchema = v.promise(); // Promise<unknown>
const RecordSchema = v.record(v.string(), v.number()); // Record<string, number>
const SetSchema = v.set(v.number()); // Set<number>
const StrictObjectSchema = v.strictObject({ key: v.string() }); // { key: string }
const StrictTupleSchema = v.strictTuple([v.string(), v.number()]); // [string, number]
const TupleSchema = v.tuple([v.string(), v.number()]); // [string, number]
const TupleWithRestSchema = v.tupleWithRest([v.string(), v.number()], v.null()); // [string, number, ...null[]]
```

#### Special cases

Beyond primitive and complex values, there are also schema functions for more special cases.

<ApiList
  label="Special schemas"
  items={[
    'any',
    'custom',
    'enum',
    'exactOptional',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'optional',
    'picklist',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

```ts
import * as v from 'valibot';

const AnySchema = v.any(); // any
const CustomSchema = v.custom<`${number}px`>(isPixelString); // `${number}px`
const EnumSchema = v.enum(Direction); // Direction
const ExactOptionalSchema = v.exactOptional(v.string()); // string
const InstanceSchema = v.instance(Error); // Error
const LazySchema = v.lazy(() => v.string()); // string
const IntersectSchema = v.intersect([v.string(), v.literal('a')]); // string & 'a'
const LiteralSchema = v.literal('foo'); // 'foo'
const NanSchema = v.nan(); // NaN
const NeverSchema = v.never(); // never
const NonNullableSchema = v.nonNullable(v.nullable(v.string())); // string
const NonNullishSchema = v.nonNullish(v.nullish(v.string())); // string
const NonOptionalSchema = v.nonOptional(v.optional(v.string())); // string
const NullableSchema = v.nullable(v.string()); // string | null
const NullishSchema = v.nullish(v.string()); // string | null | undefined
const OptionalSchema = v.optional(v.string()); // string | undefined
const PicklistSchema = v.picklist(['a', 'b']); // 'a' | 'b'
const UndefinedableSchema = v.undefinedable(v.string()); // string | undefined
const UnionSchema = v.union([v.string(), v.number()]); // string | number
const UnknownSchema = v.unknown(); // unknown
const VariantSchema = v.variant('type', [
  v.object({ type: v.literal('a'), foo: v.string() }),
  v.object({ type: v.literal('b'), bar: v.number() }),
]); // { type: 'a'; foo: string } | { type: 'b'; bar: number }
const VoidSchema = v.void(); // void
```

### Pipelines

For detailed validations and transformations, a schema can be wrapped in a pipeline. Especially for schema functions like <Link href="/api/string/">`string`</Link>, <Link href="/api/number/">`number`</Link>, <Link href="/api/date/">`date`</Link>, <Link href="/api/object/">`object`</Link>, and <Link href="/api/array/">`array`</Link>, this feature is useful for validating properties beyond the raw data type.

#### How it works

In simple words, a pipeline is a list of schemas and actions that synchronously passes through the input data. It must always start with a schema, followed by up to 19 schemas or actions. Each schema and action can examine and modify the input. The pipeline is therefore perfect for detailed validations and transformations.

##### Example

For example, the pipeline feature can be used to trim a string and make sure that it is an email that ends with a specific domain.

```ts
import * as v from 'valibot';

const EmailSchema = v.pipe(
  v.string(),
  v.trim(),
  v.email(),
  v.endsWith('@example.com')
);
```

#### Validations

Pipeline validation actions examine the input and, if the input does not meet a certain condition, return an issue. If the input is valid, it is returned as the output and, if present, picked up by the next action in the pipeline.

> Whenever possible, pipelines are run completely, even if an issue has occurred, to collect all possible issues. If you want to abort the pipeline early after the first issue, you need to set the `abortPipeEarly` option to `true`. Learn more about this <Link href="/guides/parse-data/#configuration">here</Link>.

<ApiList
  label="Validation actions"
  items={[
    'base64',
    'bic',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'finite',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'regex',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'startsWith',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

Some of these actions can be combined with different schemas. For example, <Link href="/api/minValue/">`minValue`</Link> can be used to validate the minimum value of <Link href="/api/string/">`string`</Link>, <Link href="/api/number/">`number`</Link>, <Link href="/api/bigint/">`bigint`</Link>, and <Link href="/api/date/">`date`</Link>.

```ts
import * as v from 'valibot';

const StringSchema = v.pipe(v.string(), v.minValue('foo'));
const NumberSchema = v.pipe(v.number(), v.minValue(1234));
const BigintSchema = v.pipe(v.bigint(), v.minValue(1234n));
const DateSchema = v.pipe(v.date(), v.minValue(new Date()));
```

##### Custom validation

For custom validations, <Link href="/api/check/">`check`</Link> can be used. If the function passed as the first argument returns `false`, an issue is returned. Otherwise, the input is considered valid.

```ts
import * as v from 'valibot';
import { isValidUsername } from '~/utils';

const UsernameSchema = v.pipe(
  v.string(),
  v.check(isValidUsername, 'This username is invalid.')
);
```

> You can forward the issues of a pipeline validation to a child. See the <Link href="/guides/methods/#forward">methods</Link> guide for more information.

#### Transformations

Pipeline transformation actions allow to change the value and data type of the input data. This can be useful for example to remove spaces at the beginning or end of a string or to force a minimum or maximum value.

<ApiList
  label="Transformation actions"
  items={[
    'brand',
    'filterItems',
    'findItem',
    'flavor',
    'mapItems',
    'rawTransform',
    'readonly',
    'reduceItems',
    'sortItems',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
  ]}
/>

For example, the pipeline of the following schema enforces a minimum value of 10. If the input is less than 10, it is replaced with the specified minimum value.

```ts
import * as v from 'valibot';

const NumberSchema = v.pipe(v.number(), v.toMinValue(10));
```

##### Custom transformation

For custom transformations, <Link href="/api/transform/">`transform`</Link> can be used. The function passed as the first argument is called with the input data and the return value defines the output. The following transformation changes the output of the schema to `null` for any number less than 10.

```ts
import * as v from 'valibot';

const NumberSchema = v.pipe(
  v.number(),
  v.transform((input) => (input < 10 ? null : input))
);
```

#### Metadata

In addition to the validation and transformation actions, a pipeline can also be used to add metadata to a schema. This can be useful when working with AI tools or for documentation purposes.

<ApiList
  label="Metadata actions"
  items={['description', 'metadata', 'title']}
/>

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.description(
    'A username must be between 4 and 16 characters long and can only contain letters, numbers, underscores and hyphens.'
  )
);
```

### Parse data

Now that you've learned how to create a schema, let's look at how you can use it to validate unknown data and make it type-safe. There are three different ways to do this.

> Each schema has a `~run` method. However, this is an internal API and should only be used if you know what you are doing.

#### Parse and throw

The <Link href="/api/parse/">`parse`</Link> method will throw a <Link href="/api/ValiError/">`ValiError`</Link> if the input does not match the schema. Therefore, you should use a try/catch block to catch errors. If the input matches the schema, it is valid and the output of the schema will be returned with the correct TypeScript type.

```ts
import * as v from 'valibot';

try {
  const EmailSchema = v.pipe(v.string(), v.email());
  const email = v.parse(EmailSchema, 'jane@example.com');

  // Handle errors if one occurs
} catch (error) {
  console.log(error);
}
```

#### Parse and return

If you want issues to be returned instead of thrown, you can use <Link href="/api/safeParse/">`safeParse`</Link>. The returned value then contains the `.success` property, which is `true` if the input is valid or `false` otherwise.

If the input is valid, you can use `.output` to get the output of the schema validation. Otherwise, if the input was invalid, the issues found can be accessed via `.issues`.

```ts
import * as v from 'valibot';

const EmailSchema = v.pipe(v.string(), v.email());
const result = v.safeParse(EmailSchema, 'jane@example.com');

if (result.success) {
  const email = result.output;
} else {
  console.log(result.issues);
}
```

#### Type guards

Another way to validate data that can be useful in individual cases is to use a type guard. You can use either a type predicate with the <Link href="/api/is/">`is`</Link> method or an assertion function with the <Link href="/api/assert/">`assert`</Link> method.

If a type guard is used, the issues of the validation cannot be accessed. Also, transformations have no effect and unknown keys of an object are not removed. Therefore, this approach is not as safe and powerful as the two previous ways. Also, due to a TypeScript limitation, it can currently only be used with synchronous schemas.

```ts
import * as v from 'valibot';

const EmailSchema = v.pipe(v.string(), v.email());
const data: unknown = 'jane@example.com';

if (v.is(EmailSchema, data)) {
  const email = data; // string
}
```

#### Configuration

By default, Valibot exhaustively collects every issue during validation to give you detailed feedback on why the input does not match the schema. If this is not required for your use case, you can control this behavior with `abortEarly` and `abortPipeEarly` to improve the performance of validation.

##### Abort validation

If you set `abortEarly` to `true`, data validation immediately aborts upon finding the first issue. If you just want to know if some data matches a schema, but you don't care about the details, this can improve performance.

```ts
import * as v from 'valibot';

try {
  const ProfileSchema = v.object({
    name: v.string(),
    bio: v.string(),
  });
  const profile = v.parse(
    ProfileSchema,
    { name: 'Jane', bio: '' },
    { abortEarly: true }
  );

  // Handle errors if one occurs
} catch (error) {
  console.log(error);
}
```

##### Abort pipeline

If you only set `abortPipeEarly` to `true`, the validation within a pipeline will only abort after finding the first issue. For example, if you only want to show the first error of a field when validating a form, you can use this option to improve performance.

```ts
import * as v from 'valibot';

try {
  const EmailSchema = v.pipe(v.string(), v.email(), v.endsWith('@example.com'));
  const email = v.parse(EmailSchema, 'jane@example.com', {
    abortPipeEarly: true,
  });

  // Handle errors if one occurs
} catch (error) {
  console.log(error);
}
```

### Infer types

Another cool feature of schemas is the ability to infer input and output types. This makes your work even easier because you don't have to write the type definition yourself.

#### Infer input types

The input type of a schema corresponds to the TypeScript type that the incoming data of a schema must match to be valid. To extract this type you use the utility type <Link href="/api/InferInput/">`InferInput`</Link>.

> You are probably interested in the input type only in special cases. In most cases, the output type should be sufficient.

```ts
import * as v from 'valibot';

const LoginSchema = v.object({
  email: v.string(),
  password: v.string(),
});

type LoginInput = v.InferInput<typeof LoginSchema>; // { email: string; password: string }
```

#### Infer output types

The output type differs from the input type only if you use <Link href="/api/optional/">`optional`</Link>, <Link href="/api/nullable/">`nullable`</Link>, <Link href="/api/nullish/">`nullish`</Link> or <Link href="/api/undefinedable/">`undefinedable`</Link> with a default value or <Link href="/api/brand/">`brand`</Link>, <Link href="/api/readonly/">`readonly`</Link> or <Link href="/api/transform/">`transform`</Link> to transform the input or data type of a schema after validation. The output type corresponds to the output of <Link href="/api/parse/">`parse`</Link> and <Link href="/api/safeParse/">`safeParse`</Link>. To infer it, you use the utility type <Link href="/api/InferOutput/">`InferOutput`</Link>.

```ts
import * as v from 'valibot';
import { hashPassword } from '~/utils';

const LoginSchema = v.pipe(
  v.object({
    email: v.string(),
    password: v.pipe(v.string(), v.transform(hashPassword)),
  }),
  v.transform((input) => {
    return {
      ...input,
      timestamp: new Date().toISOString(),
    };
  })
);

type LoginOutput = v.InferOutput<typeof LoginSchema>; // { email: string; password: string; timestamp: string }
```

#### Infer issue types

You can also infer the possible issues of a schema. This can be useful if you want to handle the issues in a particular way. To extract this information from a schema you use the utility type <Link href="/api/InferIssue/">`InferIssue`</Link>.

```ts
import * as v from 'valibot';

const LoginSchema = v.object({
  email: v.pipe(v.string(), v.email()),
  password: v.pipe(v.string(), v.minLength(8)),
});

type Issue = v.InferIssue<typeof LoginSchema>; // v.ObjectIssue | v.StringIssue | v.EmailIssue<string> | v.MinLengthIssue<string, 8>
```

### Methods

Apart from <Link href="/api/parse/">`parse`</Link> and <Link href="/api/safeParse/">`safeParse`</Link>, Valibot offers some more methods to make working with your schemas easier. In the following we distinguish between schema, object and pipeline methods.

#### Schema methods

Schema methods add functionality, simplify ergonomics, and help you use schemas for validation and data extraction.

<ApiList
  label="Schema methods"
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'getDefault',
    'getDefaults',
    'getDescription',
    'getFallback',
    'getFallbacks',
    'getMetadata',
    'getTitle',
    'is',
    'message',
    'parse',
    'safeParse',
    'summarize',
    'pipe',
    'unwrap',
  ]}
/>

> For more information on <Link href="/api/pipe/">`pipe`</Link>, see the <Link href="/guides/pipelines/">pipelines</Link> guide. For more information on validation methods, see the <Link href="/guides/parse-data/">parse data</Link> guide. For more information on <Link href="/api/flatten/">`flatten`</Link>, see the <Link href="/guides/issues/#formatting">issues</Link> guide.

##### Fallback

If an issue occurs while validating your schema, you can catch it with <Link href="/api/fallback/">`fallback`</Link> to return a predefined value instead.

```ts
import * as v from 'valibot';

const StringSchema = v.fallback(v.string(), 'hello');
const stringOutput = v.parse(StringSchema, 123); // 'hello'
```

#### Object methods

Object methods make it easier for you to work with object schemas. They are strongly oriented towards TypeScript's utility types.

<ApiList
  label="Object methods"
  items={['keyof', 'omit', 'partial', 'pick', 'required']}
/>

##### TypeScript similarities

Like in TypeScript, you can make the values of an object optional with <Link href="/api/partial/">`partial`</Link>, make them required with <Link href="/api/required/">`required`</Link>, and even include/exclude certain values from an existing schema with <Link href="/api/pick/">`pick`</Link> and <Link href="/api/omit/">`omit`</Link>.

```ts
import * as v from 'valibot';

// TypeScript
type Object1 = Partial<{ key1: string; key2: number }>;

// Valibot
const object1 = v.partial(v.object({ key1: v.string(), key2: v.number() }));

// TypeScript
type Object2 = Pick<Object1, 'key1'>;

// Valibot
const object2 = v.pick(object1, ['key1']);
```

#### Pipeline methods

Pipeline methods modify the results of validations and transformations within a pipeline.

<ApiList label="Pipeline methods" items={['forward']} />

> For more info about our pipeline feature, see the <Link href="/guides/pipelines/">pipelines</Link> guide.

##### Forward

‎<Link href="/api/forward/">`forward`</Link> allows you to associate an issue with a nested schema. For example, if you want to check that both password entries in a registration form match, you can use it to forward the issue to the second password field in case of an error. This allows you to display the error message in the correct place.

```ts
import * as v from 'valibot';

const RegisterSchema = v.pipe(
  v.object({
    email: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your email.'),
      v.email('The email address is badly formatted.')
    ),
    password1: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your password.'),
      v.minLength(8, 'Your password must have 8 characters or more.')
    ),
    password2: v.string(),
  }),
  v.forward(
    v.partialCheck(
      [['password1'], ['password2']],
      (input) => input.password1 === input.password2,
      'The two passwords do not match.'
    ),
    ['password2']
  )
);
```

### Issues

When validating unknown data against a schema, Valibot collects information about each issue. If there is at least one issue, these are returned in an array. Each issue provides detailed information for you or your users to fix the problem.

#### Issue info

A single issue conforms to the TypeScript type definition below.

```ts
type BaseIssue = {
  // Required info
  kind: 'schema' | 'validation' | 'transformation';
  type: string;
  input: unknown;
  expected: string | null;
  received: string;
  message: string;

  // Optional info
  requirement?: unknown;
  path?: IssuePath;
  issues?: Issues;
  lang?: string;
  abortEarly?: boolean;
  abortPipeEarly?: boolean;
  skipPipe?: boolean;
};
```

##### Required info

Each issue contains the following required information.

###### Kind

`kind` describes the kind of the problem. If an input does not match the data type, for example a number was passed instead of a string, `kind` has the value `'schema'`. In all other cases, the reason is not the data type but the actual content of the data. For example, if a string is invalid because it does not match a regex, `kind` has the value `'validation'`.

###### Type

`type` describes which function did the validation. If the schema function <Link href="/api/array/">`array`</Link> detects that the input is not an array, `type` has the value `'array'`. If the <Link href="/api/minLength/">`minLength`</Link> validation function detects that an array is too short, `type` has the value `'min_length'`.

###### Input

`input` contains the input data where the issue was found. For complex data, for example objects, `input` contains the value of the respective key that does not match the schema.

###### Expected

`expected` is a language-neutral string that describes the data property that was expected. It can be used to create useful error messages. If your users aren't developers, you can replace the language-neutral symbols with language-specific words.

###### Received

`received` is a language-neutral string that describes the data property that was received. It can be used to create useful error messages. If your users aren't developers, you can replace the language-neutral symbols with language-specific words.

###### Message

`message` contains a human-understandable error message that can be fully customized as described in our <Link href="/guides/quick-start/#error-messages">quick start</Link> and <Link href="/guides/internationalization/">internationalization</Link> guide.

##### Optional info

Some issues contain further optional information.

###### Requirement

`requirement` can contain further validation information. For example, if the <Link href="/api/minLength/">`minLength`</Link> validation function detects that a string is too short, `requirement` contains the minimum length that the string should have.

###### Path

`path` is an array of objects that describes where an issue is located within complex data. Each path item contains the following information.

> The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

```ts
type PathItem = {
  type: string;
  origin: 'key' | 'value';
  input: unknown;
  key?: unknown;
  value: unknown;
};
```

For example, you can use the following code to create a dot path.

```ts
import * as v from 'valibot';

const dotPath = v.getDotPath(issue);
```

###### Issues

`issues` currently only occur when using <Link href="/api/union/">`union`</Link> and contains all issues of the schemas of an union type.

###### Config

`lang` can be used as part of our <Link href="/guides/internationalization/">i18n feature</Link> to define the required language. `abortEarly` and `abortPipeEarly` gives you an info that the validation was aborted prematurely. You can find more info about this in the <Link href="/guides/parse-data/#configuration">parse data</Link> guide. These are all configurations that you can control yourself.

#### Formatting

For common use cases such as form validation, Valibot includes small built-in functions for formatting issues. However, once you understand how they work, you can easily format them yourself and put them in the right form for your use case.

##### Flatten errors

If you are only interested in the error messages of each issue to show them to your users, you can convert an array of issues to a flat object with <Link href="/api/flatten/">`flatten`</Link>. Below is an example.

```ts
import * as v from 'valibot';

const ObjectSchema = v.object({
  foo: v.string('Value of "foo" is missing.'),
  bar: v.object({
    baz: v.string('Value of "bar.baz" is missing.'),
  }),
});

const result = v.safeParse(ObjectSchema, { bar: {} });

if (result.issues) {
  console.log(v.flatten<typeof ObjectSchema>(result.issues));
}
```

The `result` returned in the code sample above this text contains the following issues.

```ts
[
  {
    kind: 'schema',
    type: 'string',
    input: undefined,
    expected: 'string',
    received: 'undefined',
    message: 'Value of "foo" is missing.',
    path: [
      {
        type: 'object',
        origin: 'value',
        input: {
          bar: {},
        },
        key: 'foo',
        value: undefined,
      },
    ],
  },
  {
    kind: 'schema',
    type: 'string',
    input: undefined,
    expected: 'string',
    received: 'undefined',
    message: 'Value of "bar.baz" is missing.',
    path: [
      {
        type: 'object',
        origin: 'value',
        input: {
          bar: {},
        },
        key: 'bar',
        value: {},
      },
      {
        type: 'object',
        origin: 'value',
        input: {},
        key: 'baz',
        value: undefined,
      },
    ],
  },
];
```

However, with the help of <Link href="/api/flatten/">`flatten`</Link> the issues were converted to the following object.

```ts
{
  nested: {
    foo: ['Value of "foo" is missing.'],
    'bar.baz': ['Value of "bar.baz" is missing.'],
  },
};
```

## Schemas (guides)

### Objects

To validate objects with a schema, you can use <Link href="/api/object/">`object`</Link> or <Link href="/api/record/">`record`</Link>. You use <Link href="/api/object/">`object`</Link> for an object with a specific shape and <Link href="/api/record/">`record`</Link> for objects with any number of uniform entries.

#### Object schema

The first argument is used to define the specific structure of the object. Each entry consists of a key and a schema as the value. The entries of the input are then validated against these schemas.

```ts
import * as v from 'valibot';

const ObjectSchema = v.object({
  key1: v.string(),
  key2: v.number(),
});
```

##### Loose and strict objects

The <Link href="/api/object/">`object`</Link> schema removes unknown entries. This means that entries that you have not defined in the first argument are not validated and added to the output. You can change this behavior by using the <Link href="/api/looseObject/">`looseObject`</Link> or <Link href="/api/strictObject/">`strictObject`</Link> schema instead.

The <Link href="/api/looseObject/">`looseObject`</Link> schema allows unknown entries and adds them to the output. The <Link href="/api/strictObject/">`strictObject`</Link> schema forbids unknown entries and returns an issue for the first unknown entry found.

##### Object with specific rest

Alternatively, you can also use the <Link href="/api/objectWithRest/">`objectWithRest`</Link> schema to define a specific schema for unknown entries. Any entries not defined in the first argument are then validated against the schema of the second argument.

```ts
import * as v from 'valibot';

const ObjectSchema = v.objectWithRest(
  {
    key1: v.string(),
    key2: v.number(),
  },
  v.null()
);
```

##### Pipeline validation

To validate the value of an entry based on another entry, you can wrap you schema with the <Link href="/api/check/">`check`</Link> validation action in a pipeline. You can also use <Link href="/api/forward/">`forward`</Link> to assign the issue to a specific object key in the event of an error.

> If you only want to validate specific entries, we recommend using <Link href="/api/partialCheck/">`partialCheck`</Link> instead as <Link href="/api/check/">`check`</Link> can only be executed if the input is fully typed.

```ts
import * as v from 'valibot';

const CalculationSchema = v.pipe(
  v.object({
    a: v.number(),
    b: v.number(),
    sum: v.number(),
  }),
  v.forward(
    v.check(({ a, b, sum }) => a + b === sum, 'The calculation is incorrect.'),
    ['sum']
  )
);
```

#### Record schema

For an object with any number of uniform entries, <Link href="/api/record/">`record`</Link> is the right choice. The schema passed as the first argument validates the keys of your record, and the schema passed as the second argument validates the values.

```ts
import * as v from 'valibot';

const RecordSchema = v.record(v.string(), v.number()); // Record<string, number>
```

##### Specific record keys

Instead of <Link href="/api/string/">`string`</Link>, you can also use <Link href="/api/custom/">`custom`</Link>, <Link href="/api/enum/">`enum`</Link>, <Link href="/api/literal/">`literal`</Link>, <Link href="/api/picklist/">`picklist`</Link> or <Link href="/api/union/">`union`</Link> to validate the keys.

```ts
import * as v from 'valibot';

const RecordSchema = v.record(v.picklist(['key1', 'key2']), v.number()); // { key1?: number; key2?: number }
```

Note that <Link href="/api/record/">`record`</Link> marks all literal keys as optional in this case. If you want to make them required, you can use the <Link href="/api/object/">`object`</Link> schema with the <Link href="/api/entriesFromList/">`entriesFromList`</Link> util instead.

```ts
import * as v from 'valibot';

const RecordSchema = v.object(v.entriesFromList(['key1', 'key2'], v.number())); // { key1: number; key2: number }
```

##### Pipeline validation

To validate the value of an entry based on another entry, you can wrap you schema with the <Link href="/api/check/">`check`</Link> validation action in a pipeline. You can also use <Link href="/api/forward/">`forward`</Link> to assign the issue to a specific record key in the event of an error.

```ts
import * as v from 'valibot';

const CalculationSchema = v.pipe(
  v.record(v.picklist(['a', 'b', 'sum']), v.number()),
  v.forward(
    v.check(
      ({ a, b, sum }) => (a || 0) + (b || 0) === (sum || 0),
      'The calculation is incorrect.'
    ),
    ['sum']
  )
);
```

### Arrays

To validate arrays with a schema you can use <Link href="/api/array/">`array`</Link> or <Link href="/api/tuple/">`tuple`</Link>. You use <Link href="/api/tuple/">`tuple`</Link> if your array has a specific shape and <Link href="/api/array/">`array`</Link> if it has any number of uniform items.

#### Array schema

The first argument you pass to <Link href="/api/array/">`array`</Link> is a schema, which is used to validate the items of the array.

```ts
import * as v from 'valibot';

const ArraySchema = v.array(v.number()); // number[]
```

##### Pipeline validation

To validate the length or contents of the array, you can use a pipeline.

```ts
import * as v from 'valibot';

const ArraySchema = v.pipe(
  v.array(v.string()),
  v.minLength(1),
  v.maxLength(5),
  v.includes('foo'),
  v.excludes('bar')
);
```

#### Tuple schema

A <Link href="/api/tuple/">`tuple`</Link> is an array with a specific shape. The first argument that you pass to the function is a tuple of schemas that defines its shape.

```ts
import * as v from 'valibot';

const TupleSchema = v.tuple([v.string(), v.number()]); // [string, number]
```

##### Loose and strict tuples

The <Link href="/api/tuple/">`tuple`</Link> schema removes unknown items. This means that items that you have not defined in the first argument are not validated and added to the output. You can change this behavior by using the <Link href="/api/looseTuple/">`looseTuple`</Link> or <Link href="/api/strictTuple/">`strictTuple`</Link> schema instead.

The <Link href="/api/looseTuple/">`looseTuple`</Link> schema allows unknown items and adds them to the output. The <Link href="/api/strictTuple/">`strictTuple`</Link> schema forbids unknown items and returns an issue for the first unknown item found.

##### Tuple with specific rest

Alternatively, you can also use the <Link href="/api/tupleWithRest/">`tupleWithRest`</Link> schema to define a specific schema for unknown items. Any items not defined in the first argument are then validated against the schema of the second argument.

```ts
import * as v from 'valibot';

const TupleSchema = v.tupleWithRest([v.string(), v.number()], v.null());
```

##### Pipeline validation

Similar to arrays, you can use a pipeline to validate the length and contents of a tuple.

```ts
import * as v from 'valibot';

const TupleSchema = v.pipe(
  v.tupleWithRest([v.string()], v.string()),
  v.maxLength(5),
  v.includes('foo'),
  v.excludes('bar')
);
```

### Optionals

It often happens that `undefined` or `null` should also be accepted instead of the value. To make the API more readable for this and to reduce boilerplate, Valibot offers a shortcut for this functionality with <Link href="/api/optional/">`optional`</Link>, <Link href="/api/exactOptional/">`exactOptional`</Link>, <Link href="/api/undefinedable/">`undefinedable`</Link>, <Link href="/api/nullable/">`nullable`</Link> and <Link href="/api/nullish/">`nullish`</Link>.

#### How it works

To accept `undefined` and/or `null` besides your actual value, you just have to wrap the schema in <Link href="/api/optional/">`optional`</Link>, <Link href="/api/exactOptional/">`exactOptional`</Link>, <Link href="/api/undefinedable/">`undefinedable`</Link>, <Link href="/api/nullable/">`nullable`</Link> or <Link href="/api/nullish/">`nullish`</Link>.

> Note: <Link href="/api/exactOptional/">`exactOptional`</Link> allows missing entries in objects, but does not allow `undefined` as a specified value.

```ts
import * as v from 'valibot';

const OptionalStringSchema = v.optional(v.string()); // string | undefined
const ExactOptionalStringSchema = v.exactOptional(v.string()); // string
const UndefinedableStringSchema = v.undefinedable(v.string()); // string | undefined
const NullableStringSchema = v.nullable(v.string()); // string | null
const NullishStringSchema = v.nullish(v.string()); // string | null | undefined
```

##### Use in objects

When used inside of objects, <Link href="/api/optional/">`optional`</Link>, <Link href="/api/exactOptional/">`exactOptional`</Link> and <Link href="/api/nullish/">`nullish`</Link> is a special case, as it also marks the value as optional in TypeScript with a question mark.

```ts
import * as v from 'valibot';

const OptionalKeySchema = v.object({ key: v.optional(v.string()) }); // { key?: string | undefined }
```

#### Default values

What makes <Link href="/api/optional/">`optional`</Link>, <Link href="/api/exactOptional/">`exactOptional`</Link>, <Link href="/api/undefinedable/">`undefinedable`</Link>, <Link href="/api/nullable/">`nullable`</Link> and <Link href="/api/nullish/">`nullish`</Link> unique is that the schema functions accept a default value as the second argument. Depending on the schema function, this default value is always used if the input is missing, `undefined` or `null`.

```ts
import * as v from 'valibot';

const OptionalStringSchema = v.optional(v.string(), "I'm the default!");

type OptionalStringInput = v.InferInput<typeof OptionalStringSchema>; // string | undefined
type OptionalStringOutput = v.InferOutput<typeof OptionalStringSchema>; // string
```

By providing a default value, the input type of the schema now differs from the output type. The schema in the example now accepts `string` and `undefined` as input, but returns a string as output in both cases.

##### Dynamic default values

In some cases it is necessary to generate the default value dynamically. For this purpose, a function that generates and returns the default value can also be passed as the second argument.

```ts
import * as v from 'valibot';

const NullableDateSchema = v.nullable(v.date(), () => new Date());
```

The previous example thus creates a new instance of the [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) class for each validation with `null` as input, which is then used as the default value.

##### Dependent default values

In rare cases, a default value for an optional entry may depend on the values of another entries in the same object. This can be achieved by using <Link href="/api/transform/">`transform`</Link> in the <Link href="/api/pipe/">`pipe`</Link> of the object.

```ts
import * as v from 'valibot';

const CalculationSchema = v.pipe(
  v.object({
    a: v.number(),
    b: v.number(),
    sum: v.optional(v.number()),
  }),
  v.transform((input) => ({
    ...input,
    sum: input.sum === undefined ? input.a + input.b : input.sum,
  }))
);
```

### Enums

An enumerated type is a data type that consists of a set of values. They can be represented by either an object, a TypeScript enum or, to keep things simple, an array. You use <Link href="/api/enum/">`enum`</Link> for objects and TypeScript enums and <Link href="/api/picklist/">`picklist`</Link> for arrays.

#### Enum schema

Since TypeScript enums are transpiled to JavaScript objects by the TypeScript compiler, you can use the <Link href="/api/enum/">`enum`</Link> schema function for both. Just pass your enumerated data type as the first argument to the schema function. On validation, the schema checks whether the input matches one of the values in the enum.

```ts
import * as v from 'valibot';

// As JavaScript object
const Direction = {
  Left: 'LEFT',
  Right: 'RIGHT',
} as const;

// As TypeScript enum
enum Direction {
  Left = 'LEFT',
  Right = 'RIGHT',
}

const DirectionSchema = v.enum(Direction);
```

#### Picklist schema

For a set of values represented by an array, you can use the <Link href="/api/picklist/">`picklist`</Link> schema function. Just pass your array as the first argument to the schema function. On validation, the schema checks whether the input matches one of the items in the array.

```ts
import * as v from 'valibot';

const Direction = ['LEFT', 'RIGHT'] as const;

const DirectionSchema = v.picklist(Direction);
```

##### Format array

In some cases, the array may not be in the correct format. In this case, simply use the [`.map()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/map) method to bring it into the required format.

```ts
import * as v from 'valibot';

const countries = [
  { name: 'Germany', code: 'DE' },
  { name: 'France', code: 'FR' },
  { name: 'United States', code: 'US' },
] as const;

const CountrySchema = v.picklist(countries.map((country) => country.code));
```

### Unions

An union represents a logical OR relationship. You can apply this concept to your schemas with <Link href="/api/union/">`union`</Link> and <Link href="/api/variant/">`variant`</Link>. For [discriminated unions](https://www.typescriptlang.org/docs/handbook/typescript-in-5-minutes-func.html#discriminated-unions) you use <Link href="/api/variant/">`variant`</Link> and in all other cases you use <Link href="/api/union/">`union`</Link>.

#### Union schema

The schema function <Link href="/api/union/">`union`</Link> creates an OR relationship between any number of schemas that you pass as the first argument in the form of an array. On validation, the schema returns the result of the first schema that was successfully validated.

```ts
import * as v from 'valibot';

// TypeScript
type Union = string | number;

// Valibot
const UnionSchema = v.union([v.string(), v.number()]);
```

If a bad input can be uniquely assigned to one of the schemas based on the data type, the result of that schema is returned. Otherwise, a general issue is returned that contains the issues of each schema as subissues. This is a special case within the library, as the issues of <Link href="/api/union/">`union`</Link> can contradict each other.

The following issues are returned if the input is `null` instead of a string or number. Since the input cannot be associated with a schema in this case, the issues of both schemas are returned as subissues.

```ts
[
  {
    kind: 'schema',
    type: 'union',
    input: null,
    expected: 'string | number',
    received: 'null',
    message: 'Invalid type: Expected string | number but received null',
    issues: [
      {
        kind: 'schema',
        type: 'string',
        input: null,
        expected: 'string',
        received: 'null',
        message: 'Invalid type: Expected string but received null',
      },
      {
        kind: 'schema',
        type: 'number',
        input: null,
        expected: 'number',
        received: 'null',
        message: 'Invalid type: Expected number but received null',
      },
    ],
  },
];
```

#### Variant schema

For better performance, more type safety, and a more targeted output of issues, you can use <Link href="/api/variant/">`variant`</Link> for discriminated unions. Therefore, we recommend using <Link href="/api/variant/">`variant`</Link> over <Link href="/api/union/">`union`</Link> whenever possible. A discriminated union is an OR relationship between objects that can be distinguished by a specific key.

When you call the schema function, you first specify the discriminator key. This is used to determine the schema to use for validation based on the input. The object schemas, in the form of an array, follow as the second argument.

```ts
import * as v from 'valibot';

const VariantScheme = v.variant('type', [
  v.object({
    type: v.literal('foo'),
    foo: v.string(),
  }),
  v.object({
    type: v.literal('bar'),
    bar: v.number(),
  }),
]);
```

For very complex datasets, multiple <Link href="/api/variant/">`variant`</Link> schemas can also be deeply nested within one another.

### Intersections

An intersection represents a logical AND relationship. You can apply this concept to your schemas with <Link href="/api/intersect/">`intersect`</Link> and partially by merging multiple object schemas into a new one. We recommend this approach for simple object schemas, and <Link href="/api/intersect/">`intersect`</Link> for all other cases.

#### Intersect schema

The schema function <Link href="/api/intersect/">`intersect`</Link> creates an AND relationship between any number of schemas that you pass as the first argument in the form of an array. To pass the validation, the validation of each schema passed must be successful. If this is the case, the schema merges the output of the individual schemas and returns the result. If the validation fails, the schema returns any issues that occurred.

```ts
import * as v from 'valibot';

// TypeScript
type Intersect = { foo: string } & { bar: number };

// Valibot
const IntersectSchema = v.intersect([
  v.object({ foo: v.string() }),
  v.object({ bar: v.number() }),
]);
```

#### Merge objects

Technically, there is a big difference between <Link href="/api/intersect/">`intersect`</Link> and object merging. <Link href="/api/intersect/">`intersect`</Link> is a schema function that executes the passed schemas during validation. In contrast, object merging is done during initialization to create a new object schema.

As a result, object merging usually has much better performance than <Link href="/api/intersect/">`intersect`</Link> when validating unknown data. Also, subsequent object properties overwrite the previous ones. This is not the case with <Link href="/api/intersect/">`intersect`</Link>, since the validation would fail if two properties with the same name are fundamentally different.

```ts
import * as v from 'valibot';

const ObjectSchema1 = v.object({ foo: v.string(), baz: v.number() });
const ObjectSchema2 = v.object({ bar: v.string(), baz: v.boolean() });

const MergedSchema = v.object({
  ...ObjectSchema1.entries,
  ...ObjectSchema2.entries,
}); // { foo: string; bar: string; baz: boolean }
```

In the previous code example, the `baz` property of the first object schema is overwritten by the `baz` property of the second object schema.

### Other

This guide explains other special schema functions such as <Link href="/api/literal/">`literal`</Link>, <Link href="/api/instance/">`instance`</Link>, <Link href="/api/custom/">`custom`</Link> and <Link href="/api/lazy/">`lazy`</Link> that are not covered in the other guides.

#### Literal schema

You can use <Link href="/api/literal/">`literal`</Link> to define a schema that matches a specific string, number or boolean value. Therefore, this schema is perfect for representing [literal types](https://www.typescriptlang.org/docs/handbook/2/everyday-types.html#literal-types). Usage is simple, just pass the value you want to match as the first argument.

```ts
import * as v from 'valibot';

const StringLiteralSchema = v.literal('foo'); // 'foo'
const NumberLiteralSchema = v.literal(12345); // 12345
const BooleanLiteralSchema = v.literal(true); // true
```

#### Instance schema

With schema functions like <Link href="/api/blob/">`blob`</Link>, <Link href="/api/date/">`date`</Link>, <Link href="/api/map/">`map`</Link> and <Link href="/api/set/">`set`</Link> Valibot already covers the most common JavaScript classes. However, there are many more classes that you may want to validate. For this purpose, you can use the <Link href="/api/instance/">`instance`</Link> schema function. It takes a class as its first argument and returns a schema that matches only instances of that class.

```ts
import * as v from 'valibot';

const ErrorSchema = v.instance(Error); // Error
const UrlSchema = v.instance(URL); // URL
```

#### Custom schema

The <Link href="/api/custom/">`custom`</Link> schema function is a bit more advanced. It allows you to define a schema that matches a value based on a custom function. Use it whenever you need to define a schema that cannot be expressed using any of the other schema functions.

The function receives the value to validate as its first argument and must return a boolean value. If the function returns `true`, the value is considered valid. Otherwise, it is considered invalid.

```ts
import * as v from 'valibot';

const PixelStringSchema = v.custom<`${number}px`>((input) =>
  typeof input === 'string' ? /^\d+px$/.test(input) : false
);
```

#### Lazy schema

The <Link href="/api/lazy/">`lazy`</Link> schema function allows you to define recursive schemas. A recursive schema is a schema that references itself. For example, you can use it to define a schema for a tree-like data structure.

> Due to a TypeScript limitation, the input and output types cannot be inferred automatically in this case. Therefore, you must explicitly specify these types using the <Link href="/api/GenericSchema/">`GenericSchema`</Link> type.

```ts
import * as v from 'valibot';

type BinaryTree = {
  element: string;
  left: BinaryTree | null;
  right: BinaryTree | null;
};

const BinaryTreeSchema: v.GenericSchema<BinaryTree> = v.object({
  element: v.string(),
  left: v.nullable(v.lazy(() => BinaryTreeSchema)),
  right: v.nullable(v.lazy(() => BinaryTreeSchema)),
});
```

##### JSON schema

Another practical use case for `lazy` is a schema for all possible `JSON` values. These are all values that can be serialized and deserialized using `JSON.stringify()` and `JSON.parse()`.

```ts
import * as v from 'valibot';

type JsonData =
  | string
  | number
  | boolean
  | null
  | { [key: string]: JsonData }
  | JsonData[];

const JsonSchema: v.GenericSchema<JsonData> = v.lazy(() =>
  v.union([
    v.string(),
    v.number(),
    v.boolean(),
    v.null(),
    v.record(v.string(), JsonSchema),
    v.array(JsonSchema),
  ])
);
```

## Advanced (guides)

### Naming convention

In many cases a schema is created and exported together with the inferred type. There are two naming conventions for this procedure that we recommend you to use when working with Valibot. In this guide we will explain both of them and share why we think they might make sense.

> You don't have to follow any of these conventions. They are only recommendations.

#### Convention 1

The first naming convention exports the schema and type with the same name. The advantage of this is that the names are short and the boilerplate is low, since the schema and type can be imported together.

We also recommend to follow the [PascalCase](<https://en.wikipedia.org/wiki/Naming_convention_(programming)>) naming convention. This means that each word starts with an uppercase letter. This is a common convention for TypeScript types, and since schemas basically provide runtime validation of types, it makes sense to use this convention for schemas as well.

##### Example

In the following example, a schema is created for a user object. In order to follow the naming convention, the schema and the type are exported with the same name.

```ts
import * as v from 'valibot';

export const PublicUser = v.object({
  name: v.pipe(v.string(), v.maxLength(30)),
  email: v.pipe(v.string(), v.email()),
  avatar: v.nullable(v.file()),
  bio: v.pipe(v.string(), v.maxLength(1000)),
});

export type PublicUser = v.InferOutput<typeof PublicUser>;
```

The schema and type can then be imported and used together.

```ts
import * as v from 'valibot';
import { PublicUser } from './types';

// Use `PublicUser` as a type
const publicUsers: PublicUser[] = [];

publicUsers.push(
  // Use `PublicUser` as a schema
  v.parse(PublicUser, {
    name: 'Jane Doe',
    email: 'jane@example.com',
    avatar: null,
    bio: 'Lorem ipsum ...',
  })
);
```

#### Convention 2

The first naming convention can cause naming conflicts with other classes and types. It also causes a problem when you need to export both the input and output types of a schema.

The second naming convention provides a solution. It also follows the [PascalCase](<https://en.wikipedia.org/wiki/Naming_convention_(programming)>) naming convention, but adds an appropriate suffix to each export. Schemas get the suffix `Schema`, input types get the suffix `Input` and output types get the suffix `Output`.

> If there is no difference between the input and output type, the suffix `Data` can optionally be used to indicate this.

This requires the schema and types to be imported separately, which increases the overhead. However, the naming convention is more precise, flexible, and works in any use case.

##### Example

In the following example, a schema is created for an image object. In order to follow the naming convention, the schema and the types are exported with different names.

```ts
import * as v from 'valibot';

export const ImageSchema = v.object({
  status: v.optional(v.picklist(['public', 'private']), 'private'),
  created: v.optional(v.date(), () => new Date()),
  title: v.pipe(v.string(), v.maxLength(100)),
  source: v.pipe(v.string(), v.url()),
  size: v.pipe(v.number(), v.minValue(0)),
});

export type ImageInput = v.InferInput<typeof ImageSchema>;
export type ImageOutput = v.InferOutput<typeof ImageSchema>;
```

The schema and the input and output types can then be imported and used separately.

```ts
import * as v from 'valibot';
import { ImageInput, ImageOutput, ImageSchema } from './types';

export function createImage(input: ImageInput): ImageOutput {
  return v.parse(ImageSchema, input);
}
```

> Do you have ideas for improving these conventions? We welcome your feedback and suggestions. Feel free to create an [issue](https://github.com/fabian-hiller/valibot/issues/new) on GitHub.

### Async validation

By default, Valibot validates each schema synchronously. This is usually the fastest way to validate unknown data, but sometimes you need to validate something asynchronously. For example, you might want to check if a username already exists in your database.

#### How it works

To be able to do this, Valibot provides an asynchronous implementation when necessary. The only difference is that the asynchronous implementation is promise-based. Otherwise, the API and functionality is exactly the same.

##### Naming

The asynchronous implementation starts with the same name as the synchronous one, but adds the suffix `Async` to the end. For example, the asynchronous implementation of <Link href="/api/pipe/">`pipe`</Link> is called <Link href="/api/pipeAsync/">`pipeAsync`</Link> and the asynchronous implementation of <Link href="/api/object/">`object`</Link> is called <Link href="/api/objectAsync/">`objectAsync`</Link>.

##### Nesting

Asynchronous functions can only be nested inside other asynchronous functions. This means that if you need to validate a string within an object asynchronously, you must also switch the object validation to the asynchronous implementation.

This is not necessary in the other direction. You can nest synchronous functions within asynchronous functions, and we recommend that you do so in most cases to keep complexity and bundle size to a minimum.

###### Rule of thumb

We recommend that you always start with the synchronous implementation, and only move the necessary parts to the asynchronous implementation as needed. If you are using TypeScript, it is not possible to make a mistake here, as our API is completely type-safe and will notify you when you embed an asynchronous function into a synchronous function.

##### Example

Let's say you want to validate a profile object and the username should be checked asynchronously against your database. Only the object and username validation needs to be asynchronous, the rest can stay synchronous.

```ts
import * as v from 'valibot';
import { isUsernameAvailable } from '~/api';

const ProfileSchema = v.objectAsync({
  username: v.pipeAsync(v.string(), v.checkAsync(isUsernameAvailable)),
  avatar: v.pipe(v.string(), v.url()),
  description: v.pipe(v.string(), v.maxLength(1000)),
});
```

### JSON Schema

In favor of a larger feature set and smaller bundle size, Valibot is not implemented with JSON Schema in mind. However, in some use cases, you may still need a JSON Schema. This guide will show you how to convert Valibot schemas to JSON Schema format.

#### Valibot to JSON Schema

A large part of Valibot's schemas are JSON Schema compatible and can be easily converted to the JSON Schema format using the official `toJsonSchema` function. This function is provided via a separate package called [`@valibot/to-json-schema`](https://github.com/fabian-hiller/valibot/tree/main/packages/to-json-schema).

> See the [README](https://github.com/fabian-hiller/valibot/blob/main/packages/to-json-schema/README.md) of the `@valibot/to-json-schema` package for more details. It is also recommended that you take a look at <Link href="/blog/json-schema-package-upgrade/">this blog post</Link>, which highlights recent improvements.

```ts
import { toJsonSchema } from '@valibot/to-json-schema';
import * as v from 'valibot';

const ValibotEmailSchema = v.pipe(v.string(), v.email());
const JsonEmailSchema = toJsonSchema(ValibotEmailSchema);
// -> { type: 'string', format: 'email' }
```

#### Cons of JSON Schema

Valibot schemas intentionally do not output JSON Schema natively. This is because JSON Schema is limited to JSON-compliant data structures. In addition, more advanced features like transformations are not supported. Since we want to leverage the full power of TypeScript, we output a custom format instead.

Another drawback of JSON Schema is that JSON Schema itself does not contain any validation logic. Therefore, an additional function is required that can validate the entire JSON Schema specification. This approach is usually not tree-shakable and results in a large bundle size.

In contrast, Valibot's API design and implementation is completely modular. Every schema is independent and contains its own validation logic. This allows the schemas to be plugged together like LEGO bricks, resulting in a much smaller bundle size due to tree shaking.

#### Pros of JSON Schema

Despite these drawbacks, JSON Schema is still widely used in the industry because it also has many advantages. For example, JSON Schemas can be used across programming languages and tools. In addition, JSON Schemas are serializable and can be easily stored in a database or transmitted over a network.

### Internationalization

Providing error messages in the native language of your users can improve the user experience and adoption rate of your software. That is why we offer several flexible ways to easily implement i18n.

#### Official translations

The fastest way to get started with i18n is to use Valibot's official translations. They are provided in a separate package called [`@valibot/i18n`](https://github.com/fabian-hiller/valibot/tree/main/packages/i18n).

> If you are missing a translation, feel free to open an [issue](https://github.com/fabian-hiller/valibot/issues/new) or pull request on GitHub.

##### Import translations

Each translation in this package is implemented modularly and exported as a submodule. This allows you to import only the translations you actually need to keep your bundle size small.

{/* prettier-ignore */}
```ts
// Import every translation (not recommended)
import '@valibot/i18n';

// Import every translation for a specific language
import '@valibot/i18n/de';

// Import only the translation for schema functions
import '@valibot/i18n/de/schema';

// Import only the translation for a specific pipeline function
import '@valibot/i18n/de/minLength';
```

The submodules use sideeffects to load the translations into a global storage that the schema and validation functions access when adding the error message to an issue.

##### Select language

The language used is then selected by the `lang` configuration. You can set it globally with <Link href="/api/setGlobalConfig/">`setGlobalConfig`</Link> or locally when parsing unknown data via <Link href="/api/parse/">`parse`</Link> or <Link href="/api/safeParse/">`safeParse`</Link>.

```ts
import * as v from 'valibot';

// Set the language configuration globally
v.setGlobalConfig({ lang: 'de' });

// Set the language configuration locally
v.parse(Schema, input, { lang: 'de' });
```

#### Custom translations

You can use the same APIs as [`@valibot/i18n`](https://github.com/fabian-hiller/valibot/tree/main/packages/i18n) to add your own translations to the global storage. Alternatively, you can also pass them directly to a specific schema or validation function as the first optional argument.

> You can either enter the translations manually or use an i18n library like [Paraglide JS](https://inlang.com/m/gerre34r/library-inlang-paraglideJs).

##### Set translations globally

You can add translations with <Link href="/api/setGlobalMessage/">`setGlobalMessage`</Link>, <Link href="/api/setSchemaMessage/">`setSchemaMessage`</Link> and <Link href="/api/setSpecificMessage/">`setSpecificMessage`</Link> in three different hierarchy levels. When creating an issue, I first check if a specific translation is available, then the translation for schema functions, and finally the global translation.

```ts
import * as v from 'valibot';

// Set the translation globally (can be used as a fallback)
v.setGlobalMessage((issue) => `Invalid input: ...`, 'en');

// Set the translation globally for every schema functions
v.setSchemaMessage((issue) => `Invalid type: ...`, 'en');

// Set the translation globally for a specific function
v.setSpecificMessage(v.minLength, (issue) => `Invalid length: ...`, 'en');
```

##### Set translations locally

If you prefer to define the translations individually, you can pass them as the first optional argument to schema and validation functions. We recommend using an i18n library like [Paraglide JS](https://inlang.com/m/gerre34r/library-inlang-paraglideJs) in this case.

{/* prettier-ignore */}
```ts
import * as v from 'valibot';
import * as m from './paraglide/messages.js';

const LoginSchema = v.object({
  email: v.pipe(
    v.string(),
    v.nonEmpty(m.emailRequired),
    v.email(m.emailInvalid)
  ),
  password: v.pipe(
    v.string(),
    v.nonEmpty(m.passwordRequired),
    v.minLength(8, m.passwordInvalid)
  ),
});
```

## Migration (guides)

### Migrate to v0.31.0

Migrating Valibot from an older version to v0.31.0 isn't complicated. Except for the new <Link href="/api/pipe/">`pipe`</Link> method, most things remain the same. The following guide will help you to migrate automatically or manually step by step and also point out important differences.

#### Automatic upgrade

We worked together with [Codemod](https://codemod.com/registry/valibot-migrate-to-v0-31-0) and [Grit](https://docs.grit.io/registry/github.com/fabian-hiller/valibot/migrate_to_v0_31_0) to automatically upgrade your schemas to the new version with a single CLI command. Both codemods are similar. You can use one or the other. Simply run the command in the directory of your project.

> We recommend using a version control system like [Git](https://git-scm.com/) so that you can revert changes if the codemod screws something up.

```bash
### Codemod
npx codemod valibot/migrate-to-v0.31.0

### Grit
npx @getgrit/cli apply github.com/fabian-hiller/valibot#migrate_to_v0_31_0
```

Please create an [issue](https://github.com/fabian-hiller/valibot/issues/new) if you encounter any problems or unexpected behavior with the provided codemods.

#### Restructure code

As mentioned above, one of the biggest differences is the new <Link href="/api/pipe/">`pipe`</Link> method. Previously, you passed the pipeline as an array to a schema function. Now you pass the schema with various actions to the new <Link href="/api/pipe/">`pipe`</Link> method to extend a schema.

```ts
// Change this
const Schema = v.string([v.email()]);

// To this
const Schema = v.pipe(v.string(), v.email());
```

We will be publishing a [blog post](/blog/valibot-v0.31.0-is-finally-available/) soon explaining all the benefits of this change. In the meantime, you can read the description of discussion [#463](https://github.com/fabian-hiller/valibot/discussions/463) and PR [#502](https://github.com/fabian-hiller/valibot/pull/502), which introduced this change.

#### Change names

Most of the names are the same as before. However, there are some exceptions. The following table shows all names that have changed.

| v0.30.0          | v0.31.0                                                                                                                                |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `anyAsync`       | <Link href="/api/any/">`any`</Link>                                                                                                    |
| `BaseSchema`     | <Link href="/api/GenericSchema/">`GenericSchema`</Link>                                                                                |
| `bigintAsync`    | <Link href="/api/bigint/">`bigint`</Link>                                                                                              |
| `blobAsync`      | <Link href="/api/blob/">`blob`</Link>                                                                                                  |
| `booleanAsync`   | <Link href="/api/boolean/">`boolean`</Link>                                                                                            |
| `custom`         | <Link href="/api/check/">`check`</Link>                                                                                                |
| `customAsync`    | <Link href="/api/checkAsync/">`checkAsync`</Link>                                                                                      |
| `coerce`         | <Link href="/api/pipe/">`pipe`</Link>, <Link href="/api/unknown/">`unknown`</Link> and <Link href="/api/transform/">`transform`</Link> |
| `dateAsync`      | <Link href="/api/date/">`date`</Link>                                                                                                  |
| `enumAsync`      | <Link href="/api/enum/">`enum_`</Link>                                                                                                 |
| `Input`          | <Link href="/api/InferInput/">`InferInput`</Link>                                                                                      |
| `instanceAsync`  | <Link href="/api/instance/">`instance`</Link>                                                                                          |
| `literalAsync`   | <Link href="/api/literal/">`literal`</Link>                                                                                            |
| `nanAsync`       | <Link href="/api/nan/">`nan`</Link>                                                                                                    |
| `neverAsync`     | <Link href="/api/never/">`never`</Link>                                                                                                |
| `nullAsync`      | <Link href="/api/null/">`null_`</Link>                                                                                                 |
| `numberAsync`    | <Link href="/api/number/">`number`</Link>                                                                                              |
| `Output`         | <Link href="/api/InferOutput/">`InferOutput`</Link>                                                                                    |
| `picklistAsync`  | <Link href="/api/picklist/">`picklist`</Link>                                                                                          |
| `SchemaConfig`   | <Link href="/api/Config/">`Config`</Link>                                                                                              |
| `special`        | <Link href="/api/custom/">`custom`</Link>                                                                                              |
| `specialAsync`   | <Link href="/api/customAsync/">`customAsync`</Link>                                                                                    |
| `SchemaConfig`   | <Link href="/api/string/">`Config`</Link>                                                                                              |
| `stringAsync`    | <Link href="/api/string/">`string`</Link>                                                                                              |
| `symbolAsync`    | <Link href="/api/symbol/">`symbol`</Link>                                                                                              |
| `undefinedAsync` | <Link href="/api/undefined/">`undefined_`</Link>                                                                                       |
| `unknownAsync`   | <Link href="/api/unknown/">`unknown`</Link>                                                                                            |
| `toCustom`       | <Link href="/api/transform/">`transform`</Link>                                                                                        |
| `toTrimmed`      | <Link href="/api/trim/">`trim`</Link>                                                                                                  |
| `toTrimmedEnd`   | <Link href="/api/trimEnd/">`trimEnd`</Link>                                                                                            |
| `toTrimmedStart` | <Link href="/api/trimStart/">`trimStart`</Link>                                                                                        |
| `voidAsync`      | <Link href="/api/void/">`void_`</Link>                                                                                                 |

#### Special cases

More complex schemas may require a bit more restructuring. This section provides more details on how to migrate specific functions.

##### Objects and tuples

Previously, you could pass a `rest` argument to the <Link href="/api/object/">`object`</Link> and <Link href="/api/tuple/">`tuple`</Link> schemas to define the behavior for unknown entries and items. We have removed the `rest` argument to simplify the implementation and reduce the bundle size if this functionality is not needed. If you do need this functionality, there is now a new <Link href="/api/objectWithRest/">`objectWithRest`</Link> and <Link href="/api/tupleWithRest/">`tupleWithRest`</Link> schema.

```ts
// Change this
const ObjectSchema = v.object({ key: v.string() }, v.null_());
const TupleSchema = v.tuple([v.string()], v.null_());

// To this
const ObjectSchema = v.objectWithRest({ key: v.string() }, v.null_());
const TupleSchema = v.tupleWithRest([v.string()], v.null_());
```

To further improve the developer experience, we have also added a <Link href="/api/looseObject/">`looseObject`</Link>, <Link href="/api/looseTuple/">`looseTuple`</Link>, <Link href="/api/strictObject/">`strictObject`</Link> and <Link href="/api/strictTuple/">`strictTuple`</Link> schema. These schemas allow or disallow unknown entries or items.

```ts
// Change this
const LooseObjectSchema = v.object({ key: v.string() }, v.unknown());
const LooseTupleSchema = v.tuple([v.string()], v.unknown());
const StrictObjectSchema = v.object({ key: v.string() }, v.never());
const StrictTupleSchema = v.tuple([v.string()], v.never());

// To this
const LooseObjectSchema = v.looseObject({ key: v.string() });
const LooseTupleSchema = v.looseTuple([v.string()]);
const StrictObjectSchema = v.strictObject({ key: v.string() });
const StrictTupleSchema = v.strictTuple([v.string()]);
```

##### Object merging

Since there are now 4 different object schemas, we could no longer provide a simple `merge` function that works in all cases, as we never know which schema you want to merge the other objects into. But there is a simple workaround with a similar developer experience.

```ts
const ObjectSchema1 = v.object({ foo: v.string() });
const ObjectSchema2 = v.object({ bar: v.number() });

// Change this
const MergedObject = v.merge([ObjectSchema1, ObjectSchema2]);

// To this
const MergedObject = v.object({
  ...ObjectSchema1.entries,
  ...ObjectSchema2.entries,
});
```

##### Brand and transform

Previously, <Link href="/api/brand/">`brand`</Link> and <Link href="/api/transform/">`transform`</Link> were methods that could be wrapped around a schema to modify it. With our new <Link href="/api/pipe/">`pipe`</Link> method, this is no longer necessary. Instead, <Link href="/api/brand/">`brand`</Link> and <Link href="/api/transform/">`transform`</Link> are now transformation actions that can be placed directly in a pipeline, resulting in better readability, especially for complex schemas.

```ts
// Change this
const BrandedSchema = v.brand(v.string(), 'foo');
const TransformedSchema = v.transform(v.string(), (input) => input.length);

// To this
const BrandedSchema = v.pipe(v.string(), v.brand('foo'));
const TransformedSchema = v.pipe(
  v.string(),
  v.transform((input) => input.length)
);
```

##### Coerce method

The `coerce` method has been removed because we felt it was an insecure API. In most cases, you don't want to coerce an unknown input into a specific data type. Instead, you want to transform a specific data type into another specific data type. For example, a string or a number into a date. To explicitly define the input type, we recommend using the new <Link href="/api/pipe/">`pipe`</Link> method together with the <Link href="/api/transform/">`transform`</Link> action to achieve the same functionality.

```ts
// Change this
const DateSchema = v.coerce(v.date(), (input) => new Date(input));

// To this
const DateSchema = v.pipe(
  v.union([v.string(), v.number()]),
  v.transform((input) => new Date(input))
);
```

##### Flatten issues

Previously, the <Link href="/api/flatten/">`flatten`</Link> function accepted a <Link href="/api/ValiError/">`ValiError`</Link> or an array of issues. We have simplified the implementation by only allowing an array of issues to be passed.

```ts
// Change this
const flatErrors = v.flatten(error);

// To this
const flatErrors = v.flatten(error.issues);
```

### Migrate from Zod

Migrating from [Zod](https://zod.dev/) to Valibot is very easy in most cases since both APIs have a lot of similarities. The following guide will help you migrate step by step and also point out important differences.

#### Official codemod

To make the migration as smoth as possible, we have created an official codemod that automatically migrates your Zod schemas to Valibot. Just copy your schemas into this editor and click play.

> The codemod is still in beta and may not cover all edge cases. If you encounter any problems or unexpected behaviour, please create an [issue](https://github.com/fabian-hiller/valibot/issues/new). Alternatively, you can try to fix any issues yourself and create a [pull request](https://github.com/fabian-hiller/valibot/pulls). You can find the source code [here](https://github.com/fabian-hiller/valibot/tree/main/codemod/zod-to-valibot).

<CodemodEditor />

You will soon be able to also run the codemod locally in your terminal to migrate your entire codebase at once. Stay tuned!

#### Replace imports

The first thing to do after <Link href="../installation/">installing</Link> Valibot is to update your imports. Just change your Zod imports to Valibot's and replace all occurrences of `z.` with `v.`.

```ts
// Change this
import { z } from 'zod';
const Schema = z.object({ key: z.string() });

// To this
import * as v from 'valibot';
const Schema = v.object({ key: v.string() });
```

#### Restructure code

One of the biggest differences between Zod and Valibot is the way you further validate a given type. In Zod, you chain methods like `.email` and `.endsWith`. In Valibot you use <Link href="../pipelines/">pipelines</Link> to do the same thing. This is a function that starts with a schema and is followed by up to 19 validation or transformation actions.

```ts
// Change this
const Schema = z.string().email().endsWith('@example.com');

// To this
const Schema = v.pipe(v.string(), v.email(), v.endsWith('@example.com'));
```

Due to the modular design of Valibot, also all other methods like `.parse` or `.safeParse` have to be used a little bit differently. Instead of chaining them, you usually pass the schema as the first argument and move any existing arguments one position to the right.

```ts
// Change this
const value = z.string().parse('foo');

// To this
const value = v.parse(v.string(), 'foo');
```

We recommend that you read our <Link href="../mental-model/">mental model</Link> guide to understand how the individual functions of Valibot's modular API work together.

#### Change names

Most of the names are the same as in Zod. However, there are some exceptions. The following table shows all names that have changed.

| Zod                  | Valibot                                                                                                                                     |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `and`                | <Link href="/api/intersect/">`intersect`</Link>                                                                                             |
| `catch`              | <Link href="/api/fallback/">`fallback`</Link>                                                                                               |
| `catchall`           | <Link href="/api/objectWithRest/">`objectWithRest`</Link>                                                                                   |
| `coerce`             | <Link href="/api/pipe/">`pipe`</Link>, <Link href="/api/unknown/">`unknown`</Link> and <Link href="/api/transform/">`transform`</Link>      |
| `datetime`           | <Link href="/api/isoDate/">`isoDate`</Link>, <Link href="/api/isoDateTime/">`isoDateTime`</Link>                                            |
| `default`            | <Link href="/api/optional/">`optional`</Link>                                                                                               |
| `discriminatedUnion` | <Link href="/api/variant/">`variant`</Link>                                                                                                 |
| `element`            | `item`                                                                                                                                      |
| `enum`               | <Link href="/api/picklist/">`picklist`</Link>                                                                                               |
| `extend`             | <Link href="../intersections/#merge-objects">Object merging</Link>                                                                          |
| `gt`                 | <Link href="/api/gtValue/">`gtValue`</Link>                                                                                                 |
| `gte`                | <Link href="/api/minValue/">`minValue`</Link>                                                                                               |
| `infer`              | <Link href="/api/InferOutput/">`InferOutput`</Link>                                                                                         |
| `int`                | <Link href="/api/integer/">`integer`</Link>                                                                                                 |
| `input`              | <Link href="/api/InferInput/">`InferInput`</Link>                                                                                           |
| `instanceof`         | <Link href="/api/instance/">`instance`</Link>                                                                                               |
| `intersection`       | <Link href="/api/intersect/">`intersect`</Link>                                                                                             |
| `lt`                 | <Link href="/api/ltValue/">`ltValue`</Link>                                                                                                 |
| `lte`                | <Link href="/api/maxValue/">`maxValue`</Link>                                                                                               |
| `max`                | <Link href="/api/maxLength/">`maxLength`</Link>, <Link href="/api/maxSize/">`maxSize`</Link>, <Link href="/api/maxValue/">`maxValue`</Link> |
| `min`                | <Link href="/api/minLength/">`minLength`</Link>, <Link href="/api/minSize/">`minSize`</Link>, <Link href="/api/minValue/">`minValue`</Link> |
| `nativeEnum`         | <Link href="/api/enum/">`enum`</Link>                                                                                                       |
| `negative`           | <Link href="/api/maxValue/">`maxValue`</Link>                                                                                               |
| `nonnegative`        | <Link href="/api/minValue/">`minValue`</Link>                                                                                               |
| `nonpositive`        | <Link href="/api/maxValue/">`maxValue`</Link>                                                                                               |
| `or`                 | <Link href="/api/union/">`union`</Link>                                                                                                     |
| `output`             | <Link href="/api/InferOutput/">`InferOutput`</Link>                                                                                         |
| `passthrough`        | <Link href="/api/looseObject/">`looseObject`</Link>                                                                                         |
| `positive`           | <Link href="/api/minValue/">`minValue`</Link>                                                                                               |
| `refine`             | <Link href="/api/check/">`check`</Link>, <Link href="/api/forward/">`forward`</Link>                                                        |
| `rest`               | <Link href="/api/tuple/">`tuple`</Link>                                                                                                     |
| `safe`               | <Link href="/api/safeInteger/">`safeInteger`</Link>                                                                                         |
| `shape`              | `entries`                                                                                                                                   |
| `strict`             | <Link href="/api/strictObject/">`strictObject`</Link>                                                                                       |
| `strip`              | <Link href="/api/object/">`object`</Link>                                                                                                   |
| `superRefine`        | <Link href="/api/rawCheck/">`rawCheck`</Link>, <Link href="/api/rawTransform/">`rawTransform`</Link>                                        |

#### Other details

Below are some more details that may be helpful when migrating from Zod to Valibot.

##### Object and tuple

To specify whether objects or tuples should allow or prevent unknown values, Valibot uses different schema functions. Zod uses the methods `.passthrough`, `.strict`, `.strip`, `.catchall` and `.rest` instead. See the <Link href="../objects/">objects</Link> and <Link href="../arrays/">arrays</Link> guide for more details.

```ts
// Change this
const ObjectSchema = z.object({ key: z.string() }).strict();

// To this
const ObjectSchema = v.strictObject({ key: v.string() });
```

##### Error messages

For individual error messages, you can pass a string or an object to Zod. It also allows you to differentiate between an error message for "required" and "invalid_type". With Valibot you just pass a single string instead.

```ts
// Change this
const StringSchema = z
  .string({ invalid_type_error: 'Not a string' })
  .min(5, { message: 'Too short' });

// To this
const StringSchema = v.pipe(
  v.string('Not a string'),
  v.minLength(5, 'Too short')
);
```

##### Coerce type

To enforce primitive values, you can use a method of the `coerce` object in Zod. There is no such object or function in Valibot. Instead, you use a pipeline with a <Link href="/api/transform/">`transform`</Link> action as the second argument. This forces you to explicitly define the input, resulting in safer code.

```ts
// Change this
const NumberSchema = z.coerce.number();

// To this
const NumberSchema = v.pipe(v.unknown(), v.transform(Number));
```

Instead of <Link href="/api/unknown/">`unknown`</Link> as in the previous example, we usually recommend using a specific schema such as <Link href="/api/string/">`string`</Link> to improve type safety. This allows you, for example, to validate the formatting of the string with <Link href="/api/decimal/">`decimal`</Link> before transforming it to a number.

```ts
const NumberSchema = v.pipe(v.string(), v.decimal(), v.transform(Number));
```

##### Async validation

Similar to Zod, Valibot supports synchronous and asynchronous validation. However, the API is a little bit different. See the <Link href="../async-validation/">async guide</Link> for more details.

## Schemas (API)

### any

Creates an any schema.

> This schema function exists only for completeness and is not recommended in practice. Instead, <Link href="../unknown/">`unknown`</Link> should be used to accept unknown data.

```ts
const Schema = v.any();
```

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `any`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### array

Creates an array schema.

```ts
const Schema = v.array<TItem, TMessage>(item, message);
```

#### Generics

- `TItem` <Property {...properties.TItem} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `item` <Property {...properties.item} />
- `message` <Property {...properties.message} />

##### Explanation

With `array` you can validate the data type of the input. If the input is not an array, you can use `message` to customize the error message.

> If your array has a fixed length, consider using <Link href="../tuple/">`tuple`</Link> for a more precise typing.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `array` can be used.

##### String array schema

Schema to validate an array of strings.

```ts
const StringArraySchema = v.array(v.string(), 'An array is required.');
```

##### Object array schema

Schema to validate an array of objects.

```ts
const ObjectArraySchema = v.array(v.object({ key: v.string() }));
```

##### Validate length

Schema that validates the length of an array.

```ts
const ArrayLengthSchema = v.pipe(
  v.array(v.number()),
  v.minLength(1),
  v.maxLength(3)
);
```

##### Validate content

Schema that validates the content of an array.

```ts
const ArrayContentSchema = v.pipe(
  v.array(v.string()),
  v.includes('foo'),
  v.excludes('bar')
);
```

#### Related

The following APIs can be combined with `array`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### bigint

Creates a bigint schema.

```ts
const Schema = v.bigint<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `bigint` you can validate the data type of the input. If the input is not a bigint, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `bigint` can be used.

##### Force minimum

Schema that forces a minimum bigint value.

```ts
const MinBigintSchema = v.pipe(v.bigint(), v.toMinValue(10n));
```

##### Validate maximum

Schema that validates a maximum bigint value.

```ts
const MaxBigintSchema = v.pipe(v.bigint(), v.maxValue(999n));
```

#### Related

The following APIs can be combined with `bigint`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'gtValue',
    'ltValue',
    'maxValue',
    'metadata',
    'minValue',
    'multipleOf',
    'notValue',
    'notValues',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'toMaxValue',
    'toMinValue',
    'transform',
    'value',
    'values',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### blob

Creates a blob schema.

> The `Blob` class is not available by default in Node.js v16 and below.

```ts
const Schema = v.blob<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `blob` you can validate the data type of the input. If the input is not a blob, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `blob` can be used.

##### Image schema

Schema to validate an image.

```ts
const ImageSchema = v.pipe(
  v.blob('Please select an image file.'),
  v.mimeType(['image/jpeg', 'image/png'], 'Please select a JPEG or PNG file.'),
  v.maxSize(1024 * 1024 * 10, 'Please select a file smaller than 10 MB.')
);
```

#### Related

The following APIs can be combined with `blob`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'mimeType',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### boolean

Creates a boolean schema.

```ts
const Schema = v.boolean<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `boolean` you can validate the data type of the input. If the input is not a boolean, you can use `message` to customize the error message.

> Instead of using a <Link href="../pipe/">`pipe`</Link> to force `true` or `false` as a value, in most cases it makes more sense to use <Link href="../literal/">`literal`</Link> for better typing.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `boolean` can be used.

##### Custom message

Boolean schema with a custom error message.

```ts
const BooleanSchema = v.boolean('A boolean is required');
```

#### Related

The following APIs can be combined with `boolean`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'gtValue',
    'ltValue',
    'maxValue',
    'maxWords',
    'metadata',
    'minValue',
    'notValue',
    'notValues',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'toMaxValue',
    'toMinValue',
    'transform',
    'value',
    'values',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### custom

Creates a custom schema.

> This schema function allows you to define a schema that matches a value based on a custom function. Use it whenever you need to define a schema that cannot be expressed using any of the other schema functions.

```ts
const Schema = v.custom<TInput, TMessage>(check, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `check` <Property {...properties.check} />
- `message` <Property {...properties.message} />

##### Explanation

With `custom` you can validate the data type of the input. If the input does not match the validation of `check`, you can use `message` to customize the error message.

> Make sure that the validation in `check` matches the data type of `TInput`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `custom` can be used.

##### Pixel string schema

Schema to validate a pixel string.

```ts
const PixelStringSchema = v.custom<`${number}px`>((input) =>
  typeof input === 'string' ? /^\d+px$/.test(input) : false
);
```

#### Related

The following APIs can be combined with `custom`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### date

Creates a date schema.

```ts
const Schema = v.date<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `date` you can validate the data type of the input. If the input is not a date, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `date` can be used.

##### Force minimum

Schema that forces a minimum date of today.

```ts
const MinDateSchema = v.pipe(v.date(), v.toMinValue(new Date()));
```

##### Validate range

Schema that validates a date in a range.

```ts
const DateRangeSchema = v.pipe(
  v.date(),
  v.minValue(new Date(2019, 0, 1)),
  v.maxValue(new Date(2020, 0, 1))
);
```

#### Related

The following APIs can be combined with `date`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'gtValue',
    'ltValue',
    'maxValue',
    'metadata',
    'minValue',
    'notValue',
    'notValues',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'toMaxValue',
    'toMinValue',
    'transform',
    'value',
    'values',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### enum

Creates an enum schema.

```ts
const Schema = v.enum<TEnum, TMessage>(enum, message);
```

#### Generics

- `TEnum` <Property {...properties.TEnum} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `enum` {/* prettier-ignore */}<Property {...properties.enum} />
- `message` <Property {...properties.message} />

##### Explanation

With `enum` you can validate that the input corresponds to an enum option. If the input is invalid, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `enum` can be used.

##### Direction enum

Schema to validate a direction enum option.

```ts
enum Direction {
  Left,
  Right,
}

const DirectionSchema = v.enum(Direction, 'Invalid direction');
```

#### Related

The following APIs can be combined with `enum`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### exactOptional

Creates an exact optional schema.

```ts
const Schema = v.exactOptional<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `exactOptional` the validation of your schema will pass missing object entries, and if you specify a `default_` input value, the schema will use it if the object entry is missing. For this reason, the output type may differ from the input type of the schema.

> The difference to <Link href="../optional/">`optional`</Link> is that this schema function follows the implementation of TypeScript's [`exactOptionalPropertyTypes` configuration](https://www.typescriptlang.org/tsconfig/#exactOptionalPropertyTypes) and only allows missing but not undefined object entries.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `exactOptional` can be used.

##### Exact optional object entries

Object schema with exact optional entries.

> By using a function as the `default_` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input is `undefined`.

```ts
const OptionalEntrySchema = v.object({
  key1: v.exactOptional(v.string()),
  key2: v.exactOptional(v.string(), "I'm the default!"),
  key3: v.exactOptional(v.date(), () => new Date()),
});
```

##### Unwrap exact optional schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `exactOptional`.

```ts
const OptionalNumberSchema = v.exactOptional(v.number());
const NumberSchema = v.unwrap(OptionalNumberSchema);
```

#### Related

The following APIs can be combined with `exactOptional`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### file

Creates a file schema.

> The `File` class is not available by default in Node.js v18 and below.

```ts
const Schema = v.file<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `file` you can validate the data type of the input. If the input is not a file, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `file` can be used.

##### Image schema

Schema to validate an image.

```ts
const ImageSchema = v.pipe(
  v.file('Please select an image file.'),
  v.mimeType(['image/jpeg', 'image/png'], 'Please select a JPEG or PNG file.'),
  v.maxSize(1024 * 1024 * 10, 'Please select a file smaller than 10 MB.')
);
```

#### Related

The following APIs can be combined with `file`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'mimeType',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### function

Creates a function schema.

```ts
const Schema = v.function<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `function` you can validate the data type of the input. If the input is not a function, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `function`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### instance

Creates an instance schema.

```ts
const Schema = v.instance<TClass, TMessage>(class_, message);
```

#### Generics

- `TClass` <Property {...properties.TClass} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `class_` {/* prettier-ignore */}<Property {...properties.class_} />
- `message` <Property {...properties.message} />

##### Explanation

With `instance` you can validate the data type of the input. If the input is not an instance of the specified `class_`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `instance` can be used.

##### Error schema

Schema to validate an `Error` instance.

```ts
const ErrorSchema = v.instance(Error, 'Error instance required.');
```

##### File schema

Schema to validate an `File` instance.

```ts
const FileSchema = v.pipe(
  v.instance(File),
  v.mimeType(['image/jpeg', 'image/png']),
  v.maxSize(1024 * 1024 * 10)
);
```

#### Related

The following APIs can be combined with `instance`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'gtValue',
    'ltValue',
    'maxSize',
    'maxValue',
    'metadata',
    'mimeType',
    'minSize',
    'minValue',
    'notSize',
    'notValue',
    'notValues',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'toMaxValue',
    'toMinValue',
    'transform',
    'value',
    'values',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### intersect

Creates an intersect schema.

> I recommend to read the <Link href="/guides/intersections/">intersections guide</Link> before using this schema function.

```ts
const Schema = v.intersect<TOptions, TMessage>(options, message);
```

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `intersect` you can validate if the input matches each of the given `options`. If the output of the intersection cannot be successfully merged, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `intersect` can be used.

##### Object intersection

Schema that combines two object schemas.

```ts
const ObjectSchema = v.intersect([
  v.object({ foo: v.string() }),
  v.object({ bar: v.number() }),
]);
```

#### Related

The following APIs can be combined with `intersect`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### lazy

Creates a lazy schema.

```ts
const Schema = v.lazy<TWrapped>(getter);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />

#### Parameters

- `getter` <Property {...properties.getter} />

##### Explanation

The `getter` function is called lazily to retrieve the schema. This is necessary to be able to access the input through the first argument of the `getter` function and to avoid a circular dependency for recursive schemas.

> Due to a TypeScript limitation, the input and output types of recursive schemas cannot be inferred automatically. Therefore, you must explicitly specify these types using <Link href="/api/GenericSchema/">`GenericSchema`</Link>. Please see the examples below.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `lazy` can be used.

##### Binary tree schema

Recursive schema to validate a binary tree.

```ts
type BinaryTree = {
  element: string;
  left: BinaryTree | null;
  right: BinaryTree | null;
};

const BinaryTreeSchema: v.GenericSchema<BinaryTree> = v.object({
  element: v.string(),
  left: v.nullable(v.lazy(() => BinaryTreeSchema)),
  right: v.nullable(v.lazy(() => BinaryTreeSchema)),
});
```

##### JSON data schema

Schema to validate all possible `JSON` values.

```ts
import * as v from 'valibot';

type JsonData =
  | string
  | number
  | boolean
  | null
  | { [key: string]: JsonData }
  | JsonData[];

const JsonSchema: v.GenericSchema<JsonData> = v.lazy(() =>
  v.union([
    v.string(),
    v.number(),
    v.boolean(),
    v.null(),
    v.record(v.string(), JsonSchema),
    v.array(JsonSchema),
  ])
);
```

##### Lazy union schema

Schema to validate a discriminated union of objects.

> In most cases, <Link href="/api/union/">`union`</Link> and <Link href="/api/variant/">`variant`</Link> are the better choices for creating such a schema. I recommend using `lazy` only in special cases.

```ts
const LazyUnionSchema = v.lazy((input) => {
  if (input && typeof input === 'object' && 'type' in input) {
    switch (input.type) {
      case 'email':
        return v.object({
          type: v.literal('email'),
          email: v.pipe(v.string(), v.email()),
        });
      case 'url':
        return v.object({
          type: v.literal('url'),
          url: v.pipe(v.string(), v.url()),
        });
      case 'date':
        return v.object({
          type: v.literal('date'),
          date: v.pipe(v.string(), v.isoDate()),
        });
    }
  }
  return v.never();
});
```

#### Related

The following APIs can be combined with `lazy`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'undefined',
    'union',
    'unionWithRest',
    'undefinedable',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### literal

Creates a literal schema.

```ts
const Schema = v.literal<TLiteral, TMessage>(literal, message);
```

#### Generics

- `TLiteral` <Property {...properties.TLiteral} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `literal` <Property {...properties.literal} />
- `message` <Property {...properties.message} />

##### Explanation

With `literal` you can validate that the input matches a specified value. If the input is invalid, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `literal` can be used.

##### String literal

Schema to validate a string literal.

```ts
const StringLiteralSchema = v.literal('foo');
```

##### Number literal

Schema to validate a number literal.

```ts
const NumberLiteralSchema = v.literal(26);
```

##### Boolean literal

Schema to validate a boolean literal.

```ts
const BooleanLiteralSchema = v.literal(true);
```

#### Related

The following APIs can be combined with `literal`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### looseObject

Creates a loose object schema.

```ts
const Schema = v.looseObject<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `looseObject` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object, you can use `message` to customize the error message.

> The difference to <Link href="../object/">`object`</Link> is that this schema includes any unknown entries in the output. In addition, this schema filters certain entries from the unknown entries for security reasons.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `looseObject` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### Simple object schema

Schema to validate a loose object with two specific keys.

```ts
const SimpleObjectSchema = v.looseObject({
  key1: v.string(),
  key2: v.number(),
});
```

##### Merge several objects

Schema that merges the entries of two object schemas.

```ts
const MergedObjectSchema = v.looseObject({
  ...ObjectSchema1.entries,
  ...ObjectSchema2.entries,
});
```

##### Mark keys as optional

Schema to validate an object with partial entries.

```ts
const PartialObjectSchema = v.partial(
  v.looseObject({
    key1: v.string(),
    key2: v.number(),
  })
);
```

##### Object with selected entries

Schema to validate only selected entries of a loose object.

```ts
const PickObjectSchema = v.pick(
  v.looseObject({
    key1: v.string(),
    key2: v.number(),
    key3: v.boolean(),
  }),
  ['key1', 'key3']
);
```

#### Related

The following APIs can be combined with `looseObject`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### looseTuple

Creates a loose tuple schema.

```ts
const Schema = v.looseTuple<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `looseTuple` you can validate the data type of the input and whether the content matches `items`. If the input is not an array, you can use `message` to customize the error message.

> The difference to <Link href="../tuple/">`tuple`</Link> is that this schema does include unknown items into the output.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `looseTuple` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Simple tuple schema

Schema to validate a loose tuple with two specific items.

```ts
const SimpleTupleSchema = v.looseTuple([v.string(), v.number()]);
```

#### Related

The following APIs can be combined with `looseTuple`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### map

Creates a map schema.

```ts
const Schema = v.map<TKey, TValue, TMessage>(key, value, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `map` you can validate the data type of the input and whether the entries matches `key` and `value`. If the input is not a map, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `map` can be used.

##### String map schema

Schema to validate a map with string values.

```ts
const StringMapSchema = v.map(v.string(), v.string());
```

##### Object map schema

Schema to validate a map with object values.

```ts
const ObjectMapSchema = v.map(v.string(), v.object({ key: v.string() }));
```

#### Related

The following APIs can be combined with `map`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nan

Creates a NaN schema.

```ts
const Schema = v.nan<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `nan` you can validate the data type of the input and if it is not `NaN`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `nan`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### never

Creates a never schema.

```ts
const Schema = v.never<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

When validated, `never` always returns an issue. You can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `never`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nonNullable

Creates a non nullable schema.

> This schema function can be used to override the behavior of <Link href="../nullable/">`nullable`</Link>.

```ts
const Schema = v.nonNullable<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonNullable` the validation of your schema will not pass `null` inputs. If the input is `null`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonNullable` can be used.

##### Non nullable string

Schema that does not accept `null`.

```ts
const NonNullableStringSchema = v.nonNullable(v.nullable(v.string()));
```

##### Unwrap non nullable

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nonNullable`.

```ts
const NonNullableNumberSchema = v.nonNullable(v.nullable(v.number()));
const NullableNumberSchema = v.unwrap(NonNullableNumberSchema);
```

#### Related

The following APIs can be combined with `nonNullable`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nonNullish

Creates a non nullish schema.

> This schema function can be used to override the behavior of <Link href="../nullish/">`nullish`</Link>.

```ts
const Schema = v.nonNullish<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonNullish` the validation of your schema will not pass `null` and `undefined` inputs. If the input is `null` or `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonNullish` can be used.

##### Non nullish string

Schema that does not accept `null` and `undefined`.

```ts
const NonNullishStringSchema = v.nonNullish(v.nullish(v.string()));
```

##### Unwrap non nullish

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nonNullish`.

```ts
const NonNullishNumberSchema = v.nonNullish(v.nullish(v.number()));
const NullishNumberSchema = v.unwrap(NonNullishNumberSchema);
```

#### Related

The following APIs can be combined with `nonNullish`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nonOptional

Creates a non optional schema.

> This schema function can be used to override the behavior of <Link href="../optional/">`optional`</Link>.

```ts
const Schema = v.nonOptional<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonOptional` the validation of your schema will not pass `undefined` inputs. If the input is `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonOptional` can be used.

##### Non optional string

Schema that does not accept `undefined`.

```ts
const NonOptionalStringSchema = v.nonOptional(v.optional(v.string()));
```

##### Unwrap non optional

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nonOptional`.

```ts
const NonOptionalNumberSchema = v.nonOptional(v.optional(v.number()));
const OptionalNumberSchema = v.unwrap(NonOptionalNumberSchema);
```

#### Related

The following APIs can be combined with `nonOptional`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### null

Creates a null schema.

```ts
const Schema = v.null<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `null` you can validate the data type of the input and if it is not `null`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `null`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nullable

Creates a nullable schema.

```ts
const Schema = v.nullable<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `nullable` the validation of your schema will pass `null` inputs, and if you specify a `default_` input value, the schema will use it if the input is `null`. For this reason, the output type may differ from the input type of the schema.

> Note that `nullable` does not accept `undefined` as an input. If you want to accept `undefined` inputs, use <Link href="../optional/">`optional`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullish/">`nullish`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallback/">`fallback`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nullable` can be used.

##### Nullable string schema

Schema that accepts `string` and `null`.

```ts
const NullableStringSchema = v.nullable(v.string(), "I'm the default!");
```

##### Nullable date schema

Schema that accepts [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) and `null`.

> By using a function as the `default_` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input is `null`.

```ts
const NullableDateSchema = v.nullable(v.date(), () => new Date());
```

##### Nullable entry schema

Object schema with a nullable entry.

```ts
const NullableEntrySchema = v.object({
  key: v.nullable(v.string()),
});
```

##### Unwrap nullable schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nullable`.

```ts
const NullableNumberSchema = v.nullable(v.number());
const NumberSchema = v.unwrap(NullableNumberSchema);
```

#### Related

The following APIs can be combined with `nullable`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### nullish

Creates a nullish schema.

```ts
const Schema = v.nullish<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `nullish` the validation of your schema will pass `undefined` and `null` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined` or `null`. For this reason, the output type may differ from the input type of the schema.

> Note that `nullish` accepts `undefined` and `null` as an input. If you want to accept only `null` inputs, use <Link href="../nullable/">`nullable`</Link>, and if you want to accept only `undefined` inputs, use <Link href="../optional/">`optional`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallback/">`fallback`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nullish` can be used.

##### Nullish string schema

Schema that accepts `string`, `undefined` and `null`.

```ts
const NullishStringSchema = v.nullish(v.string(), "I'm the default!");
```

##### Nullish date schema

Schema that accepts [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date), `undefined` and `null`.

> By using a function as the `default_` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input is `undefined` or `null`.

```ts
const NullishDateSchema = v.nullish(v.date(), () => new Date());
```

##### Nullish entry schema

Object schema with a nullish entry.

```ts
const NullishEntrySchema = v.object({
  key: v.nullish(v.string()),
});
```

##### Unwrap nullish schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nullish`.

```ts
const NullishNumberSchema = v.nullish(v.number());
const NumberSchema = v.unwrap(NullishNumberSchema);
```

#### Related

The following APIs can be combined with `nullish`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### number

Creates a number schema.

```ts
const Schema = v.number<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `number` you can validate the data type of the input. If the input is not a number, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `number` can be used.

##### Integer schema

Schema to validate an integer.

```ts
const IntegerSchema = v.pipe(v.number(), v.integer());
```

##### Force minimum

Schema that forces a minimum number of 10.

```ts
const MinNumberSchema = v.pipe(v.number(), v.toMinValue(10));
```

##### Validate range

Schema that validates a number in a range.

```ts
const NumberRangeSchema = v.pipe(v.number(), v.minValue(10), v.maxValue(20));
```

#### Related

The following APIs can be combined with `number`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'finite',
    'flavor',
    'gtValue',
    'integer',
    'ltValue',
    'maxValue',
    'metadata',
    'minValue',
    'multipleOf',
    'notValue',
    'notValues',
    'rawCheck',
    'rawTransform',
    'readonly',
    'safeInteger',
    'title',
    'toMaxValue',
    'toMinValue',
    'transform',
    'value',
    'values',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### object

Creates an object schema.

```ts
const Schema = v.object<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `object` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object, you can use `message` to customize the error message.

> This schema removes unknown entries. The output will only include the entries you specify. To include unknown entries, use <Link href="../looseObject/">`looseObject`</Link>. To return an issue for unknown entries, use <Link href="../strictObject/">`strictObject`</Link>. To include and validate unknown entries, use <Link href="../objectWithRest/">`objectWithRest`</Link>.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `object` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### Simple object schema

Schema to validate an object with two keys.

```ts
const SimpleObjectSchema = v.object({
  key1: v.string(),
  key2: v.number(),
});
```

##### Merge several objects

Schema that merges the entries of two object schemas.

```ts
const MergedObjectSchema = v.object({
  ...ObjectSchema1.entries,
  ...ObjectSchema2.entries,
});
```

##### Mark keys as optional

Schema to validate an object with partial entries.

```ts
const PartialObjectSchema = v.partial(
  v.object({
    key1: v.string(),
    key2: v.number(),
  })
);
```

##### Object with selected entries

Schema to validate only selected entries of an object.

```ts
const PickObjectSchema = v.pick(
  v.object({
    key1: v.string(),
    key2: v.number(),
    key3: v.boolean(),
  }),
  ['key1', 'key3']
);
```

#### Related

The following APIs can be combined with `object`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### objectWithRest

Creates an object with rest schema.

```ts
const Schema = v.objectWithRest<TEntries, TRest, TMessage>(
  entries,
  rest,
  message
);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `rest` <Property {...properties.rest} />
- `message` <Property {...properties.message} />

##### Explanation

With `objectWithRest` you can validate the data type of the input and whether the content matches `entries` and `rest`. If the input is not an object, you can use `message` to customize the error message.

> The difference to <Link href="../object/">`object`</Link> is that this schema includes unknown entries in the output. In addition, this schema filters certain entries from the unknown entries for security reasons.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `objectWithRest` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### Object schema with rest

Schema to validate an object with generic rest entries.

```ts
const ObjectSchemaWithRest = v.objectWithRest(
  {
    key1: v.string(),
    key2: v.number(),
  },
  v.boolean()
);
```

##### Merge several objects

Schema that merges the entries of two object schemas.

```ts
const MergedObjectSchema = v.objectWithRest(
  {
    ...ObjectSchema1.entries,
    ...ObjectSchema2.entries,
  },
  v.null()
);
```

##### Mark keys as optional

Schema to validate an object with partial entries.

```ts
const PartialObjectSchema = partial(
  objectWithRest(
    {
      key1: string(),
      key2: number(),
    },
    v.undefined()
  )
);
```

##### Object with selected entries

Schema to validate only selected entries of an object.

```ts
const PickObjectSchema = v.pick(
  v.objectWithRest(
    {
      key1: v.string(),
      key2: v.number(),
      key3: v.boolean(),
    },
    v.null()
  ),
  ['key1', 'key3']
);
```

#### Related

The following APIs can be combined with `objectWithRest`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### optional

Creates an optional schema.

```ts
const Schema = v.optional<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `optional` the validation of your schema will pass `undefined` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined`. For this reason, the output type may differ from the input type of the schema.

> Note that `optional` does not accept `null` as an input. If you want to accept `null` inputs, use <Link href="../nullable/">`nullable`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullish/">`nullish`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallback/">`fallback`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `optional` can be used.

##### Optional string schema

Schema that accepts `string` and `undefined`.

```ts
const OptionalStringSchema = v.optional(v.string(), "I'm the default!");
```

##### Optional date schema

Schema that accepts [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) and `undefined`.

> By using a function as the `default_` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input is `undefined`.

```ts
const OptionalDateSchema = v.optional(v.date(), () => new Date());
```

##### Optional entry schema

Object schema with an optional entry.

```ts
const OptionalEntrySchema = v.object({
  key: v.optional(v.string()),
});
```

##### Unwrap optional schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `optional`.

```ts
const OptionalNumberSchema = v.optional(v.number());
const NumberSchema = v.unwrap(OptionalNumberSchema);
```

#### Related

The following APIs can be combined with `optional`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### picklist

Creates a picklist schema.

```ts
const Schema = v.picklist<TOptions, TMessage>(options, message);
```

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `picklist` you can validate that the input corresponds to a picklist option. If the input is invalid, you can use `message` to customize the error message.

> `picklist` works in a similar way to <Link href="../enum/">`enum`</Link>. However, in many cases it is easier to use because you can pass an array of values instead of an enum.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `picklist` can be used.

##### Language schema

Schema to validate programming languages.

```ts
const LanguageSchema = v.picklist(['JavaScript', 'TypeScript']);
```

##### Country schema

Schema to validate country codes.

```ts
const countries = [
  { name: 'Germany', code: 'DE' },
  { name: 'France', code: 'FR' },
  { name: 'United States', code: 'US' },
] as const;

const CountrySchema = v.picklist(
  countries.map((country) => country.code),
  'Please select your country.'
);
```

#### Related

The following APIs can be combined with `picklist`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### promise

Creates a promise schema.

```ts
const Schema = v.promise<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `promise` you can validate the data type of the input. If the input is not a promise, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `promise` can be used.

##### Number promise

Schema to validate a promise that resolves to a number.

```ts
const NumberPromiseSchema = v.pipeAsync(
  v.promise(),
  v.awaitAsync(),
  v.number()
);
```

#### Related

The following APIs can be combined with `promise`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'awaitAsync',
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### record

Creates a record schema.

```ts
const Schema = v.record<TKey, TValue, TMessage>(key, value, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `record` you can validate the data type of the input and whether the entries matches `key` and `value`. If the input is not an object, you can use `message` to customize the error message.

> This schema filters certain entries from the record for security reasons.

> This schema marks an entry as optional if it detects that its key is a literal type. The reason for this is that it is not technically possible to detect missing literal keys without restricting the `key` schema to <Link href="../string/">`string`</Link>, <Link href="../enum/">`enum`</Link> and <Link href="../picklist/">`picklist`</Link>. However, if <Link href="../enum/">`enum`</Link> and <Link href="../picklist/">`picklist`</Link> are used, it is better to use <Link href="../object/">`object`</Link> with <Link href="../entriesFromList/">`entriesFromList`</Link> because it already covers the needed functionality. This decision also reduces the bundle size of `record`, because it only needs to check the entries of the input and not any missing keys.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `record` can be used.

##### String record schema

Schema to validate a record with strings.

```ts
const StringRecordSchema = v.record(
  v.string(),
  v.string(),
  'An object is required.'
);
```

##### Object record schema

Schema to validate a record of objects.

```ts
const ObjectRecordSchema = v.record(v.string(), v.object({ key: v.string() }));
```

##### Picklist as key

Schema to validate a record with specific optional keys.

```ts
const ProductRecordSchema = v.record(
  v.picklist(['product_a', 'product_b', 'product_c']),
  v.optional(v.number())
);
```

##### Enum as key

Schema to validate a record with specific optional keys.

```ts
enum Products {
  PRODUCT_A = 'product_a',
  PRODUCT_B = 'product_b',
  PRODUCT_C = 'product_c',
}

const ProductRecordSchema = v.record(v.enum(Products), v.optional(v.number()));
```

#### Related

The following APIs can be combined with `record`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'minEntries',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### set

Creates a set schema.

```ts
const Schema = v.set<TValue, TMessage>(value, message);
```

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `set` you can validate the data type of the input and whether the content matches `value`. If the input is not a set, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `set` can be used.

##### String set schema

Schema to validate a set with string values.

```ts
const StringSetSchema = v.set(v.string());
```

##### Object set schema

Schema to validate a set with object values.

```ts
const ObjectSetSchema = v.set(v.object({ key: v.string() }));
```

#### Related

The following APIs can be combined with `set`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### strictObject

Creates a strict object schema.

```ts
const Schema = v.strictObject<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `strictObject` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object or does include unknown entries, you can use `message` to customize the error message.

> The difference to <Link href="../object/">`object`</Link> is that this schema returns an issue for unknown entries. It intentionally returns only one issue. Otherwise, attackers could send large objects to exhaust device resources. If you want an issue for every unknown key, use the <Link href="../objectWithRest/">`objectWithRest`</Link> schema with <Link href="../never/">`never`</Link> for the `rest` argument.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `strictObject` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### Simple object schema

Schema to validate a strict object with two keys.

```ts
const SimpleObjectSchema = v.strictObject({
  key1: v.string(),
  key2: v.number(),
});
```

##### Merge several objects

Schema that merges the entries of two object schemas.

```ts
const MergedObjectSchema = v.strictObject({
  ...ObjectSchema1.entries,
  ...ObjectSchema2.entries,
});
```

##### Mark keys as optional

Schema to validate an object with partial entries.

```ts
const PartialObjectSchema = v.partial(
  v.strictObject({
    key1: v.string(),
    key2: v.number(),
  })
);
```

##### Object with selected entries

Schema to validate only selected entries of a strict object.

```ts
const PickObjectSchema = v.pick(
  v.strictObject({
    key1: v.string(),
    key2: v.number(),
    key3: v.boolean(),
  }),
  ['key1', 'key3']
);
```

#### Related

The following APIs can be combined with `strictObject`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### strictTuple

Creates a strict tuple schema.

```ts
const Schema = v.strictTuple<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `strictTuple` you can validate the data type of the input and whether the content matches `items`. If the input is not an array or does include unknown items, you can use `message` to customize the error message.

> The difference to <Link href="../tuple/">`tuple`</Link> is that this schema returns an issue for unknown items. It intentionally returns only one issue. Otherwise, attackers could send large arrays to exhaust device resources. If you want an issue for every unknown item, use the <Link href="../tupleWithRest/">`tupleWithRest`</Link> schema with <Link href="../never/">`never`</Link> for the `rest` argument.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `strictTuple` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Simple tuple schema

Schema to validate a strict tuple with two items.

```ts
const SimpleTupleSchema = v.strictTuple([v.string(), v.number()]);
```

#### Related

The following APIs can be combined with `strictTuple`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### string

Creates a string schema.

```ts
const Schema = v.string<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `string` you can validate the data type of the input. If the input is not a string, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `string` can be used.

##### Email schema

Schema to validate an email.

```ts
const EmailSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please enter your email.'),
  v.email('The email is badly formatted.'),
  v.maxLength(30, 'Your email is too long.')
);
```

##### Password schema

Schema to validate a password.

```ts
const PasswordSchema = v.pipe(
  v.string(),
  v.minLength(8, 'Your password is too short.'),
  v.maxLength(30, 'Your password is too long.'),
  v.regex(/[a-z]/, 'Your password must contain a lowercase letter.'),
  v.regex(/[A-Z]/, 'Your password must contain a uppercase letter.'),
  v.regex(/[0-9]/, 'Your password must contain a number.')
);
```

##### URL schema

Schema to validate a URL.

```ts
const UrlSchema = v.pipe(
  v.string('A URL must be string.'),
  v.url('The URL is badly formatted.')
);
```

#### Related

The following APIs can be combined with `string`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'base64',
    'bic',
    'brand',
    'bytes',
    'creditCard',
    'cuid2',
    'check',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'excludes',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'json',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'maxBytes',
    'maxGraphemes',
    'maxLength',
    'maxValue',
    'maxWords',
    'metadata',
    'minBytes',
    'minGraphemes',
    'minLength',
    'minValue',
    'minWords',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'rawCheck',
    'rawTransform',
    'readonly',
    'regex',
    'rfcEmail',
    'slug',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### symbol

Creates a symbol schema.

```ts
const Schema = v.symbol<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `symbol` you can validate the data type of the input. If it is not a symbol, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `symbol` can be used.

##### Custom message

Symbol schema with a custom error message.

```ts
const schema = v.symbol('A symbol is required');
```

#### Related

The following APIs can be combined with `symbol`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### tuple

Creates a tuple schema.

```ts
const Schema = v.tuple<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `tuple` you can validate the data type of the input and whether the content matches `items`. If the input is not an array, you can use `message` to customize the error message.

> This schema removes unknown items. The output will only include the items you specify. To include unknown items, use <Link href="../looseTuple/">`looseTuple`</Link>. To return an issue for unknown items, use <Link href="../strictTuple/">`strictTuple`</Link>. To include and validate unknown items, use <Link href="../tupleWithRest/">`tupleWithRest`</Link>.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `tuple` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Simple tuple schema

Schema to validate a tuple with two items.

```ts
const SimpleTupleSchema = v.tuple([v.string(), v.number()]);
```

#### Related

The following APIs can be combined with `tuple`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### tupleWithRest

Creates a tuple with rest schema.

```ts
const Schema = v.tupleWithRest<TItems, TRest, TMessage>(items, rest, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `rest` <Property {...properties.rest} />
- `message` <Property {...properties.message} />

##### Explanation

With `tupleWithRest` you can validate the data type of the input and whether the content matches `items` and `rest`. If the input is not an array, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `tupleWithRest` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Tuple schema with rest

Schema to validate a tuple with generic rest items.

```ts
const TupleSchemaWithRest = v.tupleWithRest(
  [v.string(), v.number()],
  v.boolean()
);
```

#### Related

The following APIs can be combined with `tupleWithRest`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### undefined

Creates an undefined schema.

```ts
const Schema = v.undefined<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `undefined` you can validate the data type of the input and if it is not `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `undefined`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### undefinedable

Creates an undefinedable schema.

```ts
const Schema = v.undefinedable<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `undefinedable` the validation of your schema will pass `undefined` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined`. For this reason, the output type may differ from the input type of the schema.

> `undefinedable` behaves exactly the same as <Link href="../optional/">`optional`</Link> at runtime. The only difference is the input and output type when used for object entries. While <Link href="../optional/">`optional`</Link> adds a question mark to the key, `undefinedable` does not.

> Note that `undefinedable` does not accept `null` as an input. If you want to accept `null` inputs, use <Link href="../nullable/">`nullable`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullish/">`nullish`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallback/">`fallback`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `undefinedable` can be used.

##### Undefinedable string schema

Schema that accepts `string` and `undefined`.

```ts
const UndefinedableStringSchema = v.undefinedable(
  v.string(),
  "I'm the default!"
);
```

##### Undefinedable date schema

Schema that accepts [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) and `undefined`.

> By using a function as the `default_` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input is `undefined`.

```ts
const UndefinedableDateSchema = v.undefinedable(v.date(), () => new Date());
```

##### Undefinedable entry schema

Object schema with an undefinedable entry.

```ts
const UndefinedableEntrySchema = v.object({
  key: v.undefinedable(v.string()),
});
```

##### Unwrap undefinedable schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `undefinedable`.

```ts
const UndefinedableNumberSchema = v.undefinedable(v.number());
const NumberSchema = v.unwrap(UndefinedableNumberSchema);
```

#### Related

The following APIs can be combined with `undefinedable`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonUndefinedable',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### union

Creates an union schema.

> I recommend that you read the <Link href="/guides/unions/">unions guide</Link> before using this schema function.

```ts
const Schema = v.union<TOptions, TMessage>(options, message);
```

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `union` you can validate if the input matches one of the given `options`. If the input does not match a schema and cannot be clearly assigned to one of the options, you can use `message` to customize the error message.

If a bad input can be uniquely assigned to one of the schemas based on the data type, the result of that schema is returned. Otherwise, a general issue is returned that contains the issues of each schema as subissues. This is a special case within the library, as the issues of `union` can contradict each other.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `union` can be used.

##### URL schema

Schema to validate an URL or empty string.

```ts
const UrlSchema = v.union([v.pipe(v.string(), v.url()), v.literal('')]);
```

##### Number schema

Schema to validate a number or decimal string.

```ts
const NumberSchema = v.union([v.number(), v.pipe(v.string(), v.decimal())]);
```

##### Date schema

Schema to validate a `Date` or ISO timestamp.

```ts
const DateSchema = v.union([v.date(), v.pipe(v.string(), v.isoTimestamp())]);
```

#### Related

The following APIs can be combined with `union`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### unknown

Creates an unknown schema.

> Use this schema function only if the data is truly unknown. Otherwise, use the other more specific schema functions that describe the data exactly.

```ts
const Schema = v.unknown();
```

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `unknown`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### variant

Creates a variant schema.

```ts
const Schema = v.variant<TKey, TOptions, TMessage>(key, options, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `variant` you can validate if the input matches one of the given object `options`. The object schema to be used for the validation is determined by the discriminator `key`. If the input does not match a schema and cannot be clearly assigned to one of the options, you can use `message` to customize the error message.

> It is allowed to specify the exact same or a similar discriminator multiple times. However, in such cases `variant` will only return the output of the first untyped or typed variant option result. Typed results take precedence over untyped ones.

> For deeply nested `variant` schemas with several different discriminator keys, `variant` will return an issue for the first most likely object schemas on invalid input. The order of the discriminator keys and the presence of a discriminator in the input are taken into account.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `variant` can be used.

##### Variant schema

Schema to validate an email, URL or date variant.

```ts
const VariantSchema = v.variant('type', [
  v.object({
    type: v.literal('email'),
    email: v.pipe(v.string(), v.email()),
  }),
  v.object({
    type: v.literal('url'),
    url: v.pipe(v.string(), v.url()),
  }),
  v.object({
    type: v.literal('date'),
    date: v.pipe(v.string(), v.isoDate()),
  }),
]);
```

##### Nested variant schema

You can also nest `variant` schemas.

```ts
const NestedVariantSchema = v.variant('type', [
  VariantSchema,
  v.object({
    type: v.literal('color'),
    date: v.pipe(v.string(), v.hexColor()),
  }),
]);
```

##### Complex variant schema

You can also use `variant` to validate complex objects with multiple different discriminator keys.

```ts
const ComplexVariantSchema = v.variant('kind', [
  v.variant('type', [
    v.object({
      kind: v.literal('fruit'),
      type: v.literal('apple'),
      item: v.object({ … }),
    }),
    v.object({
      kind: v.literal('fruit'),
      type: v.literal('banana'),
      item: v.object({ … }),
    }),
  ]),
  v.variant('type', [
    v.object({
      kind: v.literal('vegetable'),
      type: v.literal('carrot'),
      item: v.object({ … }),
    }),
    v.object({
      kind: v.literal('vegetable'),
      type: v.literal('tomato'),
      item: v.object({ … }),
    }),
  ]),
]);
```

#### Related

The following APIs can be combined with `variant`.

##### Schemas

<ApiList items={['object']} />

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

### void

Creates a void schema.

```ts
const Schema = v.void<TMessage>(message);
```

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `void` you can validate the data type of the input and if it is not `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Related

The following APIs can be combined with `void`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

## Methods (API)

### assert

Checks if the input matches the scheme.

> As this is an assertion function, it can be used as a type guard.

```ts
v.assert<TSchema>(schema, input);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />

##### Explanation

`assert` does not modify the `input`. Therefore, transformations have no effect and unknown keys of an object are not removed. That is why this approach is not as safe and powerful as <Link href='../parse/'>`parse`</Link> and <Link href='../safeParse/'>`safeParse`</Link>.

#### Example

The following example show how `assert` can be used.

```ts
const EmailSchema = v.pipe(v.string(), v.email());
const data: unknown = 'jane@example.com';

v.assert(EmailSchema, data);
const email = data; // string
```

#### Related

The following APIs can be combined with `assert`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

### config

Changes the local configuration of a schema.

```ts
const Schema = v.config<TSchema>(schema, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `config` <Property {...properties.config} />

##### Explanation

This method overwrites the selected configuration properties by merging the previous configuration of the `schema` with the provided `config`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `config` can be used.

##### Same error message

Schema that uses the same error message for the entire pipeline.

```ts
const Schema = v.object({
  email: v.config(
    v.pipe(v.string(), v.trim(), v.email(), v.endsWith('@example.com')),
    { message: 'The email does not conform to the required format.' }
  ),
  // ...
});
```

##### Abort pipeline early

Schema that aborts only a specific pipeline early.

```ts
const Schema = v.object({
  url: v.config(
    v.pipe(v.string(), v.trim(), v.url(), v.endsWith('@example.com')),
    { abortPipeEarly: true }
  ),
  // ...
});
```

#### Related

The following APIs can be combined with `config`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### fallback

Returns a fallback value as output if the input does not match the schema.

```ts
const Schema = v.fallback<TSchema, TFallback>(schema, fallback);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TFallback` <Property {...properties.TFallback} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `fallback` <Property {...properties.fallback} />

##### Explanation

`fallback` allows you to define a fallback value for the output that will be used if the validation of the input fails. This means that no issues will be returned when using `fallback` and the schema will always return an output.

> If you only want to set a default value for `null` or `undefined` inputs, you should use <Link href="../optional/">`optional`</Link>, <Link href="../nullable/">`nullable`</Link> or <Link href="../nullish/">`nullish`</Link> instead.

> The fallback value is not validated. Make sure that the fallback value matches your schema.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `fallback` can be used.

##### Fallback string schema

Schema that will always return a string output.

```ts
const FallbackStringSchema = v.fallback(v.string(), "I'm the fallback!");
```

##### Fallback date schema

Schema that will always return a [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) output.

> By using a function as the `fallback` parameter, the schema will return a new [`Date`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) instance each time the input does not match the schema.

```ts
const FallbackDateSchema = v.fallback(v.date(), () => new Date());
```

#### Related

The following APIs can be combined with `fallback`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### flatten

Flatten the error messages of issues.

```ts
const errors = v.flatten<TSchema>(issues);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `issues` <Property {...properties.issues} />

##### Explanation

The error messages of issues without a path that belong to the root of the schema are added to the `.root` key.

The error messages of issues with a path that belong to the nested parts of the schema and can be converted to a dot path are added to the `.nested` key.

Some issue paths, for example for complex data types like `Set` and `Map`, have no key or a key that cannot be converted to a dot path. These error messages are added to the `.other` key.

#### Returns

- `errors` <Property {...properties.errors} />

#### Examples

The following example show how `flatten` can be used.

```ts
const Schema = v.object({
  nested: v.object({
    foo: v.string('Value of "nested.foo" is invalid.'),
  }),
});

const result = v.safeParse(Schema, { nested: { foo: null } });

if (result.issues) {
  const flatErrors = v.flatten<typeof Schema>(result.issues);

  // ...
}
```

#### Related

The following APIs can be combined with `flatten`.

##### Methods

<ApiList items={['parse', 'parser', 'safeParse']} />

### forward

Forwards the issues of the passed validation action.

```ts
const Action = v.forward<TInput, TIssue, TPath>(action, path);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TIssue` <Property {...properties.TIssue} />
- `TPath` <Property {...properties.TPath} />

#### Parameters

- `action` <Property {...properties.action} />
- `path` <Property {...properties.path} />

##### Explanation

`forward` allows you to forward the issues of the passed validation `action` via `path` to a nested field of a schema.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `forward` can be used.

##### Register schema

Schema that ensures that the two passwords match.

```ts
const RegisterSchema = v.pipe(
  v.object({
    email: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your email.'),
      v.email('The email address is badly formatted.')
    ),
    password1: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your password.'),
      v.minLength(8, 'Your password must have 8 characters or more.')
    ),
    password2: v.string(),
  }),
  v.forward(
    v.partialCheck(
      [['password1'], ['password2']],
      (input) => input.password1 === input.password2,
      'The two passwords do not match.'
    ),
    ['password2']
  )
);
```

#### Related

The following APIs can be combined with `forward`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'custom',
    'looseObject',
    'looseTuple',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'union',
    'unknown',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['omit', 'partial', 'pick', 'pipe', 'required']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'reduceItems',
    'someItem',
    'sortItems',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### getDefault

Returns the default value of the schema.

```ts
const value = v.getDefault<TSchema>(schema, dataset, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `dataset` <Property {...properties.dataset} />
- `config` <Property {...properties.config} />

#### Returns

- `value` <Property {...properties.value} />

#### Examples

The following examples show how `getDefault` can be used.

##### Optional string schema

Get the default value of an optional string schema.

```ts
const OptionalStringSchema = v.optional(v.string(), "I'm the default!");
const defaultValue = v.getDefault(OptionalStringSchema); // "I'm the default!"
```

#### Related

The following APIs can be combined with `getDefault`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### getDefaults

Returns the default values of the schema.

> The difference to <Link href='../getDefault/'>`getDefault`</Link> is that for object and tuple schemas this function recursively returns the default values of the subschemas instead of `undefined`.

```ts
const values = v.getDefaults<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `values` <Property {...properties.values} />

#### Examples

The following examples show how `getDefaults` can be used.

##### Object defaults

Get the default values of an object schema.

```ts
const ObjectSchema = v.object({
  key: v.optional(v.string(), "I'm the default!"),
});

const defaultValues = v.getDefaults(ObjectSchema); // { key: "I'm the default!" }
```

##### Tuple defaults

Get the default values of a tuple schema.

```ts
const TupleSchema = v.tuple([v.nullable(v.number(), 100)]);
const defaultValues = v.getDefaults(TupleSchema); // [100]
```

#### Related

The following APIs can be combined with `getDefaults`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

### getDescription

Returns the description of the schema.

> If multiple descriptions are defined, the last one of the highest level is returned. If no description is defined, `undefined` is returned.

```ts
const description = v.getDescription<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `description` <Property {...properties.description} />

#### Examples

The following examples show how `getDescription` can be used.

##### Get description of schema

Get the description of a username schema.

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.description(
    'A username must be between 4 and 16 characters long and can only contain letters, numbers, underscores and hyphens.'
  )
);

const description = v.getDescription(UsernameSchema);
```

##### Overriding inherited descriptions

Get the description of a Gmail schema with an overridden description.

```ts
const EmailSchema = v.pipe(v.string(), v.email(), v.description('Email'));

const GmailSchema = v.pipe(
  EmailSchema,
  v.endsWith('@gmail.com'),
  v.description('Gmail')
);

const description = v.getDescription(GmailSchema); // 'Gmail'
```

#### Related

The following APIs can be combined with `getDescription`.

##### Actions

<ApiList items={['description']} />

### getFallback

Returns the fallback value of the schema.

```ts
const value = v.getFallback<TSchema>(schema, dataset, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `dataset` <Property {...properties.dataset} />
- `config` <Property {...properties.config} />

#### Returns

- `value` <Property {...properties.value} />

#### Examples

The following examples show how `getFallback` can be used.

##### Fallback string schema

Get the fallback value of a string schema.

```ts
const FallbackStringSchema = v.fallback(v.string(), "I'm the fallback!");
const fallbackValue = v.getFallback(FallbackStringSchema); // "I'm the fallback!"
```

#### Related

The following APIs can be combined with `getFallback`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### getFallbacks

Returns the fallback values of the schema.

> The difference to <Link href='../getFallback/'>`getFallback`</Link> is that for object and tuple schemas this function recursively returns the fallback values of the subschemas instead of `undefined`.

```ts
const values = v.getFallbacks<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `values` <Property {...properties.values} />

#### Examples

The following examples show how `getFallbacks` can be used.

##### Object fallbacks

Get the fallback values of an object schema.

```ts
const ObjectSchema = v.object({
  key: v.fallback(v.string(), "I'm the fallback!"),
});

const fallbackValues = v.getFallbacks(ObjectSchema); // { key: "I'm the fallback!" }
```

##### Tuple fallbacks

Get the fallback values of a tuple schema.

```ts
const TupleSchema = v.tuple([v.fallback(v.number(), 100)]);
const fallbackValues = v.getFallbacks(TupleSchema); // [100]
```

#### Related

The following APIs can be combined with `getFallbacks`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

### getMetadata

Returns the metadata of the schema.

> If multiple metadata are defined, it shallowly merges them using depth-first search. If no metadata is defined, an empty object is returned.

```ts
const metadata = v.getMetadata<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `metadata` <Property {...properties.metadata} />

#### Examples

The following examples show how `getMetadata` can be used.

##### Get metadata of schema

Get the metadata of a username schema.

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.metadata({
    length: { min: 4, max: 16 },
    chars: ['letters', 'numbers', 'underscores', 'hyphens'],
  })
);

const metadata = v.getMetadata(UsernameSchema);
```

#### Related

The following APIs can be combined with `getMetadata`.

##### Actions

<ApiList items={['metadata']} />

### getTitle

Returns the title of the schema.

> If multiple titles are defined, the last one of the highest level is returned. If no title is defined, `undefined` is returned.

```ts
const title = v.getTitle<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `title` <Property {...properties.title} />

#### Examples

The following examples show how `getTitle` can be used.

##### Get title of schema

Get the title of a username schema.

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.description(
    'A username must be between 4 and 16 characters long and can only contain letters, numbers, underscores and hyphens.'
  )
);

const title = v.getTitle(UsernameSchema); // 'Username'
```

##### Overriding inherited titles

Get the title of a Gmail schema with an overridden title.

```ts
const EmailSchema = v.pipe(v.string(), v.email(), v.title('Email'));

const GmailSchema = v.pipe(
  EmailSchema,
  v.endsWith('@gmail.com'),
  v.title('Gmail')
);

const title = v.getTitle(GmailSchema); // 'Gmail'
```

#### Related

The following APIs can be combined with `getTitle`.

##### Actions

<ApiList items={['title']} />

### is

Checks if the input matches the scheme.

> By using a type predicate, this function can be used as a type guard.

```ts
const result = v.is<TSchema>(schema, input);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />

##### Explanation

`is` does not modify the `input`. Therefore, transformations have no effect and unknown keys of an object are not removed. That is why this approach is not as safe and powerful as <Link href='../parse/'>`parse`</Link> and <Link href='../safeParse/'>`safeParse`</Link>.

#### Returns

- `result` <Property {...properties.result} />

#### Example

The following example show how `is` can be used.

```ts
const EmailSchema = v.pipe(v.string(), v.email());
const data: unknown = 'jane@example.com';

if (v.is(EmailSchema, data)) {
  const email = data; // string
}
```

#### Related

The following APIs can be combined with `is`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

### keyof

Creates a picklist schema of object keys.

```ts
const Schema = v.keyof<TSchema, TMessage>(schema, message);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `message` <Property {...properties.message} />

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `keyof` can be used.

##### Object key schema

Schema to validate the keys of an object.

```ts
const ObjectSchema = v.object({ key1: v.string(), key2: v.number() });
const ObjectKeySchema = v.keyof(ObjectSchema); // 'key1' | 'key2'
```

#### Related

The following APIs can be combined with `keyof`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'metadata',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### message

Changes the local message configuration of a schema.

```ts
const Schema = v.message<TSchema>(schema, message_);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `message_` <Property {...properties['message_']} />

##### Explanation

This method overrides the local message configuration of the schema. In practice, it is typically used to specify a single error message for an entire pipeline.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `message` can be used.

##### Email schema

Email schema that uses the same error message for the entire pipeline.

```ts
const EmailSchema = v.message(
  v.pipe(v.string(), v.trim(), v.nonEmpty(), v.email(), v.maxLength(100)),
  'The email is not in the required format.'
);
```

#### Related

The following APIs can be combined with `message`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'pipe',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### omit

Creates a modified copy of an object schema that does not contain the selected entries.

```ts
const Schema = v.omit<TSchema, TKeys>(schema, keys);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />

##### Explanation

`omit` creates a modified copy of the given object `schema` that does not contain the selected `keys`. It is similar to TypeScript's [`Omit`](https://www.typescriptlang.org/docs/handbook/utility-types.html#omittype-keys) utility type.

> Because `omit` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipe/'>`pipe`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipe/'>`pipe`</Link> method after you have modified the schema with `omit`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `omit` can be used.

##### Omit specific keys

Schema that does not contain the selected keys of an existing schema.

```ts
const OmittedSchema = v.omit(
  v.object({
    key1: v.string(),
    key2: v.number(),
    key3: v.boolean(),
  }),
  ['key1', 'key3']
); // { key2: number }
```

#### Related

The following APIs can be combined with `omit`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'parse',
    'parser',
    'partial',
    'pick',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
  ]}
/>

### parse

Parses an unknown input based on a schema.

```ts
const output = v.parse<TSchema>(schema, input, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />
- `config` <Property {...properties.config} />

##### Explanation

`parse` will throw a <Link href="../ValiError/">`ValiError`</Link> if the `input` does not match the `schema`. Therefore you should use a try/catch block to catch errors. If the input matches the schema, it is valid and the `output` of the schema will be returned typed.

#### Returns

- `output` <Property {...properties.output} />

#### Example

The following example show how `parse` can be used.

```ts
try {
  const EmailSchema = v.pipe(v.string(), v.email());
  const email = v.parse(EmailSchema, 'jane@example.com');

  // Handle errors if one occurs
} catch (error) {
  console.log(error);
}
```

#### Related

The following APIs can be combined with `parse`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath', 'isValiError', 'ValiError']} />

### parser

Returns a function that parses an unknown input based on a schema.

```ts
const parser = v.parser<TSchema, TConfig>(schema, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `config` <Property {...properties.config} />

#### Returns

- `parser` <Property {...properties.parser} />

#### Example

The following example show how `parser` can be used.

```ts
try {
  const EmailSchema = v.pipe(v.string(), v.email());
  const emailParser = v.parser(EmailSchema);
  const email = emailParser('jane@example.com');

  // Handle errors if one occurs
} catch (error) {
  console.log(error);
}
```

#### Related

The following APIs can be combined with `parser`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath', 'isValiError', 'ValiError']} />

### partial

Creates a modified copy of an object schema that marks all or only the selected entries as optional.

```ts
const Schema = v.partial<TSchema, TKeys>(schema, keys);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />

##### Explanation

`partial` creates a modified copy of the given object `schema` where all entries or only the selected `keys` are optional. It is similar to TypeScript's [`Partial`](https://www.typescriptlang.org/docs/handbook/utility-types.html#partialtype) utility type.

> Because `partial` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipe/'>`pipe`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipe/'>`pipe`</Link> method after you have modified the schema with `partial`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `partial` can be used.

##### Partial object schema

Schema to validate an object with partial entries.

```ts
const PartialSchema = v.partial(
  v.object({
    key1: v.string(),
    key2: v.number(),
  })
); // { key1?: string; key2?: number }
```

##### With only specific keys

Schema to validate an object with only specific entries marked as optional.

```ts
const PartialSchema = v.partial(
  v.object({
    key1: v.string(),
    key2: v.number(),
    key3: v.boolean(),
  }),
  ['key1', 'key3']
); // { key1?: string; key2: number; key3?: boolean }
```

#### Related

The following APIs can be combined with `partial`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'pick',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### pick

Creates a modified copy of an object schema that contains only the selected entries.

```ts
const Schema = v.pick<TSchema, TKeys>(schema, keys);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />

##### Explanation

`pick` creates a modified copy of the given object `schema` that contains only the selected `keys`. It is similar to TypeScript's [`Pick`](https://www.typescriptlang.org/docs/handbook/utility-types.html#picktype-keys) utility type.

> Because `pick` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipe/'>`pipe`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipe/'>`pipe`</Link> method after you have modified the schema with `pick`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `pick` can be used.

##### Pick specific keys

Schema that contains only the selected keys of an existing schema.

```ts
const PickedSchema = v.pick(
  object({
    key1: string(),
    key2: number(),
    key3: boolean(),
  }),
  ['key1', 'key3']
); // { key1: string; key3: boolean }
```

#### Related

The following APIs can be combined with `pick`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
  ]}
/>

### pipe

Adds a pipeline to a schema, that can validate and transform its input.

```ts
const Schema = v.pipe<TSchema, TItems>(schema, ...items);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TItems` <Property {...properties.TItems} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `items` <Property {...properties.items} />

##### Explanation

`pipe` creates a modified copy of the given `schema`, containing a pipeline for detailed validations and transformations. It passes the input data synchronously through the `items` in the order they are provided and each item can examine and modify it.

> Since `pipe` returns a schema that can be used as the first argument of another pipeline, it is possible to nest multiple `pipe` calls to extend the validation and transformation further.

The `pipe` aborts early and marks the output as untyped if issues were collected before attempting to execute a schema or transformation action as the next item in the pipeline, to prevent unexpected behavior.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `pipe` can be used. Please see the <Link href="/guides/pipelines/">pipeline guide</Link> for more examples and explanations.

##### Email schema

Schema to validate an email.

```ts
const EmailSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please enter your email.'),
  v.email('The email is badly formatted.'),
  v.maxLength(30, 'Your email is too long.')
);
```

##### String to number

Schema to convert a string to a number.

```ts
const NumberSchema = v.pipe(v.string(), v.transform(Number), v.number());
```

#### Related

The following APIs can be combined with `pipe`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### required

Creates a modified copy of an object schema that marks all or only the selected entries as required.

```ts
const AllKeysSchema = v.required<TSchema, TMessage>(schema, message);
const SelectedKeysSchema = v.required<TSchema, TKeys, TMessage>(
  schema,
  keys,
  message
);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />
- `message` <Property {...properties.message} />

##### Explanation

`required` creates a modified copy of the given object `schema` where all or only the selected `keys` are required. It is similar to TypeScript's [`Required`](https://www.typescriptlang.org/docs/handbook/utility-types.html#requiredtype) utility type.

> Because `required` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipe/'>`pipe`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipe/'>`pipe`</Link> method after you have modified the schema with `required`.

#### Returns

- `AllKeysSchema` <Property {...properties.AllKeysSchema} />
- `SelectedKeysSchema` <Property {...properties.SelectedKeysSchema} />

#### Examples

The following examples show how `required` can be used.

##### Required object schema

Schema to validate an object with required entries.

```ts
const RequiredSchema = v.required(
  v.object({
    key1: v.optional(v.string()),
    key2: v.optional(v.number()),
  })
); // { key1: string; key2: number }
```

##### With only specific keys

Schema to validate an object with only specific entries marked as required.

```ts
const RequiredSchema = v.required(
  v.object({
    key1: v.optional(v.string()),
    key2: v.optional(v.number()),
    key3: v.optional(v.boolean()),
  }),
  ['key1', 'key3']
); // { key1: string; key2?: number; key3: boolean }
```

#### Related

The following APIs can be combined with `required`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### safeParse

Parses an unknown input based on a schema.

```ts
const result = v.safeParse<TSchema>(schema, input, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />
- `config` <Property {...properties.config} />

#### Returns

- `result` <Property {...properties.result} />

#### Example

The following example show how `safeParse` can be used.

```ts
const EmailSchema = v.pipe(v.string(), v.email());
const result = v.safeParse(EmailSchema, 'jane@example.com');

if (result.success) {
  const email = result.output;
} else {
  console.log(result.issues);
}
```

#### Related

The following APIs can be combined with `safeParse`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath']} />

### safeParser

Returns a function that parses an unknown input based on a schema.

```ts
const safeParser = v.safeParser<TSchema, TConfig>(schema, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `config` <Property {...properties.config} />

#### Returns

- `safeParser` <Property {...properties.safeParser} />

#### Example

The following example show how `safeParser` can be used.

```ts
const EmailSchema = v.pipe(v.string(), v.email());
const safeEmailParser = v.safeParser(EmailSchema);
const result = safeEmailParser('jane@example.com');

if (result.success) {
  const email = result.output;
} else {
  console.log(result.issues);
}
```

#### Related

The following APIs can be combined with `safeParser`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath']} />

### summarize

Summarize the error messages of issues in a pretty-printable multi-line string.

```ts
const errors = v.summarize(issues);
```

#### Parameters

- `issues` <Property {...properties.issues} />

##### Explanation

If an issue in `issues` contains a path that can be converted to a dot path, the dot path will be displayed in the `errors` output just below the issue's error message.

#### Returns

- `errors` <Property {...properties.errors} />

#### Examples

The following example show how `summarize` can be used.

```ts
const Schema = v.object({
  nested: v.object({
    foo: v.string('Value of "nested.foo" is invalid.'),
  }),
});

const result = v.safeParse(Schema, { nested: { foo: null } });

if (result.issues) {
  console.log(v.summarize(result.issues));
}
```

#### Related

The following APIs can be combined with `summarize`.

##### Methods

<ApiList items={['parse', 'parser', 'safeParse']} />

### unwrap

Unwraps the wrapped schema.

```ts
const Schema = v.unwrap<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `unwrap` can be used.

##### Unwrap string schema

Unwraps the wrapped string schema.

```ts
const OptionalStringSchema = v.optional(v.string());
const StringSchema = v.unwrap(OptionalStringSchema);
```

#### Related

The following APIs can be combined with `unwrap`.

##### Schemas

<ApiList
  items={[
    'exactOptional',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'optional',
    'undefinedable',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'is',
    'message',
    'parse',
    'parser',
    'pipe',
    'safeParse',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'exactOptionalAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'optionalAsync',
    'undefinedableAsync',
  ]}
/>

## Actions (API)

### args

Creates a function arguments transformation action.

```ts
const Action = v.args<TInput, TSchema>(schema);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

##### Explanation

With `args` you can force the arguments of a function to match the given `schema`.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `args` can be used.

##### Function schema

Schema of a function that transforms a string to a number.

```ts
const FunctionSchema = v.pipe(
  v.function(),
  v.args(v.tuple([v.pipe(v.string(), v.decimal())])),
  v.returns(v.number())
);
```

#### Related

The following APIs can be combined with `args`.

##### Schemas

<ApiList
  items={[
    'any',
    'custom',
    'looseTuple',
    'function',
    'strictTuple',
    'tuple',
    'tupleWithRest',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### base64

Creates a [Base64](https://en.wikipedia.org/wiki/Base64) validation action.

```ts
const Action = v.base64<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `base64` you can validate the formatting of a string. If the input is not a Base64 string, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `base64` can be used.

##### Base64 schema

Schema to validate a Base64 string.

```ts
const Base64Schema = v.pipe(v.string(), v.base64('The data is badly encoded.'));
```

#### Related

The following APIs can be combined with `base64`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### bic

Creates a [BIC](https://en.wikipedia.org/wiki/ISO_9362) validation action.

```ts
const Action = v.bic<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `bic` you can validate the formatting of a string. If the input is not a BIC, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `bic` can be used.

##### BIC schema

Schema to validate a BIC.

```ts
const BicSchema = v.pipe(
  v.string(),
  v.toUpperCase(),
  v.bic('The BIC is badly formatted.')
);
```

#### Related

The following APIs can be combined with `bic`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### brand

Creates a brand transformation action.

```ts
const Action = v.brand<TInput, TName>(name);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TName` <Property {...properties.TName} />

#### Parameters

- `name` <Property {...properties.name} />

##### Explanation

`brand` allows you to brand the output type of a schema with a `name`. This ensures that data can only be considered valid if it has been validated by a particular branded schema.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `brand` can be used.

##### Branded fruit schema

Schema to ensure that only a validated fruit is accepted.

```ts
// Create schema and infer output type
const FruitSchema = v.pipe(v.object({ name: v.string() }), v.brand('Fruit'));
type FruitOutput = v.InferOutput<typeof FruitSchema>;

// This works because output is branded
const apple: FruitOutput = v.parse(FruitSchema, { name: 'apple' });

// But this will result in a type error
const banana: FruitOutput = { name: 'banana' };
```

#### Related

The following APIs can be combined with `brand`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### bytes

Creates a [bytes](https://en.wikipedia.org/wiki/Byte) validation action.

```ts
const Action = v.bytes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `bytes` you can validate the bytes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `bytes` can be used.

##### Bytes schema

Schema to validate a string with 8 bytes.

```ts
const BytesSchema = v.pipe(
  v.string(),
  v.bytes(8, 'Exactly 8 bytes are required.')
);
```

#### Related

The following APIs can be combined with `bytes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### check

Creates a check validation action.

```ts
const Action = v.check<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `check` you can freely validate the input and return `true` if it is valid or `false` otherwise. If the input does not match your `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `check` can be used.

##### Check object properties

Schema to check the properties of an object.

```ts
const CustomObjectSchema = v.pipe(
  v.object({
    list: v.array(v.string()),
    length: v.number(),
  }),
  v.check(
    (input) => input.list.length === input.length,
    'The list does not match the length.'
  )
);
```

#### Related

The following APIs can be combined with `check`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['forward', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### checkItems

Creates a check items validation action.

```ts
const Action = v.checkItems<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `checkItems` you can freely validate the items of an array and return `true` if they are valid or `false` otherwise. If an item does not match your `requirement`, you can use `message` to customize the error message.

> The special thing about `checkItems` is that it automatically forwards each issue to the appropriate item.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `checkItems` can be used.

##### No duplicate items

Schema to validate that an array has no duplicate items.

```ts
const ArraySchema = v.pipe(
  v.array(v.string()),
  v.checkItems(
    (item, index, array) => array.indexOf(item) === index,
    'Duplicate items are not allowed.'
  )
);
```

#### Related

The following APIs can be combined with `checkItems`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### creditCard

Creates a [credit card](https://en.wikipedia.org/wiki/Payment_card_number) validation action.

```ts
const Action = v.creditCard<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `creditCard` you can validate the formatting of a string. If the input is not a credit card, you can use `message` to customize the error message.

> The following credit card providers are currently supported: American Express, Diners Card, Discover, JCB, Union Pay, Master Card, and Visa.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `creditCard` can be used.

##### Credit Card schema

Schema to validate a credit card.

```ts
const CreditCardSchema = v.pipe(
  v.string(),
  v.creditCard('The credit card is badly formatted.')
);
```

#### Related

The following APIs can be combined with `creditCard`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### cuid2

Creates a [Cuid2](https://github.com/paralleldrive/cuid2) validation action.

```ts
const Action = v.cuid2<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `cuid2` you can validate the formatting of a string. If the input is not an Cuid2, you can use `message` to customize the error message.

> Since Cuid2s are not limited to a fixed length, it is recommended to combine `cuid2` with <Link href="../length/">`length`</Link> to ensure the correct length.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `cuid2` can be used.

##### Cuid2 schema

Schema to validate an Cuid2.

```ts
const Cuid2Schema = v.pipe(
  v.string(),
  v.cuid2('The Cuid2 is badly formatted.'),
  v.length(10, 'The Cuid2 must be 10 characters long.')
);
```

#### Related

The following APIs can be combined with `cuid2`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### decimal

Creates a [decimal](https://en.wikipedia.org/wiki/Decimal) validation action.

> The difference between `decimal` and <Link href="../digits/">`digits`</Link> is that `decimal` accepts floating point numbers and negative numbers, while <Link href="../digits/">`digits`</Link> accepts only the digits 0-9.

```ts
const Action = v.decimal<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `decimal` you can validate the formatting of a string. If the input is not a decimal, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `decimal` can be used.

##### Decimal schema

Schema to validate a decimal.

```ts
const DecimalSchema = v.pipe(
  v.string(),
  v.decimal('The decimal is badly formatted.')
);
```

#### Related

The following APIs can be combined with `decimal`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### description

Creates a description metadata action.

```ts
const Action = v.description<TInput, TDescription>(description_);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TDescription` <Property {...properties.TDescription} />

#### Parameters

- `description_` <Property {...properties['description_']} />

##### Explanation

With `description` you can describe the purpose of a schema. This can be useful when working with AI tools or for documentation purposes.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `description` can be used.

##### Username schema

Schema to validate a user name.

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.description(
    'A username must be between 4 and 16 characters long and can only contain letters, numbers, underscores and hyphens.'
  )
);
```

#### Related

The following APIs can be combined with `description`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['getDescription', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### digits

Creates a [digits](https://en.wikipedia.org/wiki/Numerical_digit) validation action.

> The difference between `digits` and <Link href="../decimal/">`decimal`</Link> is that `digits` accepts only the digits 0-9, while <Link href="../decimal/">`decimal`</Link> accepts floating point numbers and negative numbers.

```ts
const Action = v.digits<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `digits` you can validate the formatting of a string. If the input does not soley consist of numerical digits, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `digits` can be used.

##### Digits schema

Schema to validate a digits.

```ts
const DigitsSchema = v.pipe(
  v.string(),
  v.digits('The string contains something other than digits.')
);
```

#### Related

The following APIs can be combined with `digits`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### email

Creates an [email](https://en.wikipedia.org/wiki/Email_address) validation action.

```ts
const Action = v.email<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `email` you can validate the formatting of a string. If the input is not an email, you can use `message` to customize the error message.

> This validation action intentionally only validates common email addresses. If you are interested in an action that covers the entire specification, please use the <Link href="../rfcEmail/">`rfcEmail`</Link> action instead.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `email` can be used.

##### Email schema

Schema to validate an email.

```ts
const EmailSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please enter your email.'),
  v.email('The email is badly formatted.'),
  v.maxLength(30, 'Your email is too long.')
);
```

#### Related

The following APIs can be combined with `email`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### emoji

Creates an [emoji](https://en.wikipedia.org/wiki/Emoji) validation action.

```ts
const Action = v.emoji<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `emoji` you can validate the formatting of a string. If the input is not an emoji, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `emoji` can be used.

##### Emoji schema

Schema to validate an emoji.

```ts
const EmojiSchema = v.pipe(
  v.string(),
  v.emoji('Please provide a valid emoji.')
);
```

#### Related

The following APIs can be combined with `emoji`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### empty

Creates an empty validation action.

```ts
const Action = v.empty<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `empty` you can validate that a string or array is empty. If the input is not empty, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `empty` can be used.

##### String schema

Schema to validate that a string is empty.

```ts
const StringSchema = v.pipe(v.string(), v.empty('The string must be empty.'));
```

##### Array schema

Schema to validate that an array is empty.

```ts
const ArraySchema = v.pipe(
  v.array(v.number()),
  v.empty('The array must be empty.')
);
```

#### Related

The following APIs can be combined with `empty`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### endsWith

Creates an ends with validation action.

```ts
const Action = v.endsWith<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `endsWith` you can validate the end of a string. If the end does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `endsWith` can be used.

##### Email schema

Schema to validate an email with a specific domain.

```ts
const EmailSchema = v.pipe(v.string(), v.email(), v.endsWith('@example.com'));
```

#### Related

The following APIs can be combined with `endsWith`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### entries

Creates an entries validation action.

```ts
const Action = v.entries<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `entries` you can validate the number of entries of an object. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `entries` can be used.

##### Exact object entries

Schema to validate an object that does have 5 entries.

```ts
const EntriesSchema = v.pipe(
  v.record(v.string(), v.number()),
  v.entries(5, 'Object must have 5 entries')
);
```

#### Related

The following APIs can be combined with `entries`.

##### Schemas

<ApiList
  items={[
    'looseObject',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### everyItem

Creates an every item validation action.

```ts
const Action = v.everyItem<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `everyItem` you can freely validate the items of an array and return `true` if they are valid or `false` otherwise. If not every item matches your `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `everyItem` can be used.

##### Sorted array schema

Schema to validate that an array is sorted.

```ts
const SortedArraySchema = v.pipe(
  v.array(v.number()),
  v.everyItem(
    (item, index, array) => index === 0 || item >= array[index - 1],
    'The numbers must be sorted in ascending order.'
  )
);
```

#### Related

The following APIs can be combined with `everyItem`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### excludes

Creates an excludes validation action.

```ts
const Action = v.excludes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `excludes` you can validate the content of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `excludes` can be used.

##### String schema

Schema to validate that a string does not contain a specific substring.

```ts
const StringSchema = v.pipe(
  v.string(),
  v.excludes('foo', 'The string must not contain "foo".')
);
```

##### Array schema

Schema to validate that an array does not contain a specific string.

```ts
const ArraySchema = v.pipe(
  v.array(v.string()),
  v.excludes('foo', 'The array must not contain "foo".')
);
```

#### Related

The following APIs can be combined with `excludes`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### filterItems

Creates a filter items transformation action.

```ts
const Action = v.filterItems<TInput>(operation);
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Parameters

- `operation` <Property {...properties.operation} />

##### Explanation

With `filterItems` you can filter the items of an array. Returning `true` for an item will keep it in the array and returning `false` will remove it.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `filterItems` can be used.

##### Filter duplicate items

Schema to filter duplicate items from an array.

```ts
const FilteredArraySchema = v.pipe(
  v.array(v.string()),
  v.filterItems((item, index, array) => array.indexOf(item) === index)
);
```

#### Related

The following APIs can be combined with `filterItems`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### findItem

Creates a find item transformation action.

```ts
const Action = v.findItem<TInput>(operation);
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Parameters

- `operation` <Property {...properties.operation} />

##### Explanation

With `findItem` you can extract the first item of an array that matches the given `operation`.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `findItem` can be used.

##### Find duplicate item

Schema to find the first duplicate item in an array.

```ts
const DuplicateItemSchema = v.pipe(
  v.array(v.string()),
  v.findItem((item, index, array) => array.indexOf(item) !== index)
);
```

#### Related

The following APIs can be combined with `findItem`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### finite

Creates a [finite](https://en.wikipedia.org/wiki/Finite) validation action.

```ts
const Action = v.finite<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `finite` you can validate the value of a number. If the input is not a finite number, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `finite` can be used.

##### Finite number schema

Schema to validate a finite number.

```ts
const FiniteNumberSchema = v.pipe(
  v.number(),
  v.finite('The number must be finite.')
);
```

#### Related

The following APIs can be combined with `finite`.

##### Schemas

<ApiList items={['any', 'custom', 'number', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### flavor

Creates a flavor transformation action.

```ts
const Action = v.flavor<TInput, TName>(name);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TName` <Property {...properties.TName} />

#### Parameters

- `name` <Property {...properties.name} />

##### Explanation

`flavor` is a less strict version of <Link href='../brand/'>`brand`</Link> that allows you to flavor the output type of a schema with a `name`. Data is considered valid if it's type is unflavored or has been validated by a schema that has the same flavor.

> `flavor` can also be used as a TypeScript DX hack to improve the editor's autocompletion by displaying only literal types, but still allowing the unflavored root type to be passed.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `flavor` can be used.

##### Flavored ID schemas

Schema to ensure that different types of IDs are not mixed up.

```ts
// Create user ID and order ID schema
const UserIdSchema = v.pipe(v.string(), v.flavor('UserId'));
const OrderIdSchema = v.pipe(v.string(), v.flavor('OrderId'));

// Infer output types of both schemas
type UserId = v.InferOutput<typeof UserIdSchema>;
type OrderId = v.InferOutput<typeof OrderIdSchema>;

// This works because output is flavored
const userId: UserId = v.parse(UserIdSchema, 'c28443ef...');
const orderId: OrderId = v.parse(OrderIdSchema, '4b717520...');

// You can also use unflavored strings
const newUserId1: UserId = '2d80cd94...';

// But this will result in a type error
const newUserId2: UserId = orderId;
```

#### Related

The following APIs can be combined with `flavor`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### graphemes

Creates a [graphemes](https://en.wikipedia.org/wiki/Grapheme) validation action.

```ts
const Action = v.graphemes<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `graphemes` you can validate the graphemes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `graphemes` can be used.

##### Graphemes schema

Schema to validate a string with 8 graphemes.

```ts
const GraphemesSchema = v.pipe(
  v.string(),
  v.graphemes(8, 'Exactly 8 graphemes are required.')
);
```

#### Related

The following APIs can be combined with `graphemes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### gtValue

Creates a greater than value validation action.

```ts
const Action = v.gtValue<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `gtValue` you can validate the value of a string, number, boolean or date. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `gtValue` can be used.

##### Number schema

Schema to validate a number with a greater than value.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.gtValue(100, 'The number must be greater than 100.')
);
```

##### Date schema

Schema to validate a date with a greater than year.

```ts
const DateSchema = v.pipe(
  v.date(),
  v.gtValue(
    new Date('2000-01-01'),
    'The date must be greater than 1st January 2000.'
  )
);
```

#### Related

The following APIs can be combined with `gtValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### hash

Creates a [hash](https://en.wikipedia.org/wiki/Hash_function) validation action.

```ts
const Action = v.hash<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `types` <Property {...properties.types} />
- `message` <Property {...properties.message} />

##### Explanation

With `hash` you can validate the formatting of a string. If the input is not a hash, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `hash` can be used.

##### Hash schema

Schema to validate a hash.

```ts
const HashSchema = v.pipe(
  v.string(),
  v.hash(['md5', 'sha1'], 'The specified hash is invalid.')
);
```

#### Related

The following APIs can be combined with `hash`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### hexadecimal

Creates a [hexadecimal](https://en.wikipedia.org/wiki/Hexadecimal) validation action.

```ts
const Action = v.hexadecimal<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `hexadecimal` you can validate the formatting of a string. If the input is not a hexadecimal, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `hexadecimal` can be used.

##### Hexadecimal schema

Schema to validate a Hexadecimal string.

```ts
const HexadecimalSchema = v.pipe(
  v.string(),
  v.hexadecimal('The hexadecimal is badly formatted.')
);
```

#### Related

The following APIs can be combined with `hexadecimal`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### hexColor

Creates a [hex color](https://en.wikipedia.org/wiki/Web_colors#Hex_triplet) validation action.

```ts
const Action = v.hexColor<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `hexColor` you can validate the formatting of a string. If the input is not a hex color, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `hexColor` can be used.

##### Hex color schema

Schema to validate a hex color.

```ts
const HexColorSchema = v.pipe(
  v.string(),
  v.hexColor('The hex color is badly formatted.')
);
```

#### Related

The following APIs can be combined with `hexColor`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### imei

Creates an [IMEI](https://en.wikipedia.org/wiki/International_Mobile_Equipment_Identity) validation action.

```ts
const Action = v.imei<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `imei` you can validate the formatting of a string. If the input is not an imei, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `imei` can be used.

##### IMEI schema

Schema to validate an IMEI.

```ts
const ImeiSchema = v.pipe(v.string(), v.imei('The imei is badly formatted.'));
```

#### Related

The following APIs can be combined with `imei`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### includes

Creates an includes validation action.

```ts
const Action = v.includes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `includes` you can validate the content of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `includes` can be used.

##### String schema

Schema to validate that a string contains a specific substring.

```ts
const StringSchema = v.pipe(
  v.string(),
  v.includes('foo', 'The string must contain "foo".')
);
```

##### Array schema

Schema to validate that an array contains a specific string.

```ts
const ArraySchema = v.pipe(
  v.array(v.string()),
  v.includes('foo', 'The array must contain "foo".')
);
```

#### Related

The following APIs can be combined with `includes`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### integer

Creates an [integer](https://en.wikipedia.org/wiki/Integer) validation action.

```ts
const Action = v.integer<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `integer` you can validate the value of a number. If the input is not an integer, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `integer` can be used.

##### Integer schema

Schema to validate an integer.

```ts
const IntegerSchema = v.pipe(
  v.number(),
  v.integer('The number must be an integer.')
);
```

#### Related

The following APIs can be combined with `integer`.

##### Schemas

<ApiList items={['any', 'custom', 'number', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### ip

Creates an [IP address](https://en.wikipedia.org/wiki/IP_address) validation action.

> This validation action accepts IPv4 and IPv6 addresses. For a more specific validation, you can also use <Link href="../ipv4/">`ipv4`</Link> or <Link href="../ipv6/">`ipv6`</Link>.

```ts
const Action = v.ip<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `ip` you can validate the formatting of a string. If the input is not an IP address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `ip` can be used.

##### IP address schema

Schema to validate an IP address.

```ts
const IpAddressSchema = v.pipe(
  v.string(),
  v.ip('The IP address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `ip`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### ipv4

Creates an [IPv4](https://en.wikipedia.org/wiki/IPv4) address validation action.

```ts
const Action = v.ipv4<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `ipv4` you can validate the formatting of a string. If the input is not an IPv4 address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `ipv4` can be used.

##### IPv4 schema

Schema to validate an IPv4 address.

```ts
const Ipv4Schema = v.pipe(
  v.string(),
  v.ipv4('The IP address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `ipv4`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### ipv6

Creates an [IPv6](https://en.wikipedia.org/wiki/IPv6) address validation action.

```ts
const Action = v.ipv6<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `ipv6` you can validate the formatting of a string. If the input is not an IPv6 address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `ipv6` can be used.

##### IPv6 schema

Schema to validate an IPv6 address.

```ts
const Ipv6Schema = v.pipe(
  v.string(),
  v.ipv6('The IP address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `ipv6`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoDate

Creates an [ISO date](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Format: `yyyy-mm-dd`

> The regex used cannot validate the maximum number of days based on year and month. For example, "2023-06-31" is valid although June has only 30 days.

```ts
const Action = v.isoDate<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoDate` you can validate the formatting of a string. If the input is not an ISO date, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoDate` can be used.

##### ISO date schema

Schema to validate an ISO date.

```ts
const IsoDateSchema = v.pipe(
  v.string(),
  v.isoDate('The date is badly formatted.')
);
```

##### Minimum value schema

Schema to validate an ISO date is after a certain date.

```ts
const MinValueSchema = v.pipe(
  v.string(),
  v.isoDate(),
  v.minValue('2000-01-01', 'The date must be after the year 1999.')
);
```

#### Related

The following APIs can be combined with `isoDate`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoDateTime

Creates an [ISO date time](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Format: `yyyy-mm-ddThh:mm`

> The regex used cannot validate the maximum number of days based on year and month. For example, "2023-06-31T00:00" is valid although June has only 30 days.

> The regex also allows a space as a separator between the date and time parts instead of the "T" character.

```ts
const Action = v.isoDateTime<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoDateTime` you can validate the formatting of a string. If the input is not an ISO date time, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoDateTime` can be used.

##### ISO date time schema

Schema to validate an ISO date time.

```ts
const IsoDateTimeSchema = v.pipe(
  v.string(),
  v.isoDateTime('The date is badly formatted.')
);
```

#### Related

The following APIs can be combined with `isoDateTime`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoTime

Creates an [ISO time](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Format: `hh:mm`

```ts
const Action = v.isoTime<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoTime` you can validate the formatting of a string. If the input is not an ISO time, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoTime` can be used.

##### ISO time schema

Schema to validate an ISO time.

```ts
const IsoTimeSchema = v.pipe(
  v.string(),
  v.isoTime('The time is badly formatted.')
);
```

#### Related

The following APIs can be combined with `isoTime`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoTimeSecond

Creates an [ISO time second](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Format: `hh:mm:ss`

```ts
const Action = v.isoTimeSecond<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoTimeSecond` you can validate the formatting of a string. If the input is not an ISO time second, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoTimeSecond` can be used.

##### ISO time second schema

Schema to validate an ISO time second.

```ts
const IsoTimeSecondSchema = v.pipe(
  v.string(),
  v.isoTimeSecond('The time is badly formatted.')
);
```

#### Related

The following APIs can be combined with `isoTimeSecond`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoTimestamp

Creates an [ISO timestamp](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Formats: `yyyy-mm-ddThh:mm:ss.sssZ`, `yyyy-mm-ddThh:mm:ss.sss±hh:mm`, `yyyy-mm-ddThh:mm:ss.sss±hhmm`

> To support timestamps with lower or higher accuracy, the millisecond specification can be removed or contain up to 9 digits.

> The regex used cannot validate the maximum number of days based on year and month. For example, "2023-06-31T00:00:00.000Z" is valid although June has only 30 days.

> The regex also allows a space as a separator between the date and time parts instead of the "T" character.

```ts
const Action = v.isoTimestamp<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoTimestamp` you can validate the formatting of a string. If the input is not an ISO timestamp, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoTimestamp` can be used.

##### ISO timestamp schema

Schema to validate an ISO timestamp.

```ts
const IsoTimestampSchema = v.pipe(
  v.string(),
  v.isoTimestamp('The timestamp is badly formatted.')
);
```

#### Related

The following APIs can be combined with `isoTimestamp`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### isoWeek

Creates an [ISO week](https://en.wikipedia.org/wiki/ISO_8601) validation action.

Format: `yyyy-Www`

> The regex used cannot validate the maximum number of weeks based on the year. For example, "2021W53" is valid although 2021 has only 52 weeks.

```ts
const Action = v.isoWeek<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `isoWeek` you can validate the formatting of a string. If the input is not an ISO week, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `isoWeek` can be used.

##### ISO week schema

Schema to validate an ISO week.

```ts
const IsoWeekSchema = v.pipe(
  v.string(),
  v.isoWeek('The week is badly formatted.')
);
```

#### Related

The following APIs can be combined with `isoWeek`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### length

Creates a length validation action.

```ts
const Action = v.length<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `length` you can validate the length of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `length` can be used.

##### String schema

Schema to validate the length of a string.

```ts
const StringSchema = v.pipe(
  v.string(),
  v.length(8, 'The string must be 8 characters long.')
);
```

##### Array schema

Schema to validate the length of an array.

```ts
const ArraySchema = v.pipe(
  v.array(v.number()),
  v.length(100, 'The array must contain 100 numbers.')
);
```

#### Related

The following APIs can be combined with `length`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### mac

Creates a [MAC address](https://en.wikipedia.org/wiki/MAC_address) validation action.

> This validation action accepts 48-bit and 64-bit MAC addresses. For a more specific validation, you can also use <Link href="../mac48/">`mac48`</Link> or <Link href="../mac64/">`mac64`</Link>.

```ts
const Action = v.mac<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `mac` you can validate the formatting of a string. If the input is not a MAC address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `mac` can be used.

##### MAC schema

Schema to validate a MAC address.

```ts
const MacSchema = v.pipe(
  v.string(),
  v.mac('The MAC address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `mac`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### mac48

Creates a 48-bit [MAC address](https://en.wikipedia.org/wiki/MAC_address) validation action.

```ts
const Action = v.mac48<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `mac48` you can validate the formatting of a string. If the input is not a 48-bit MAC address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `mac48` can be used.

##### 48-bit MAC schema

Schema to validate a 48-bit MAC address.

```ts
const Mac48Schema = v.pipe(
  v.string(),
  v.mac48('The MAC address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `mac48`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### mac64

Creates a 64-bit [MAC address](https://en.wikipedia.org/wiki/MAC_address) validation action.

```ts
const Action = v.mac64<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `mac64` you can validate the formatting of a string. If the input is not a 64-bit MAC address, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `mac64` can be used.

##### 64-bit MAC schema

Schema to validate a 64-bit MAC address.

```ts
const Mac64Schema = v.pipe(
  v.string(),
  v.mac64('The MAC address is badly formatted.')
);
```

#### Related

The following APIs can be combined with `mac64`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### mapItems

Creates a map items transformation action.

```ts
const Action = v.mapItems<TInput, TOutput>(operation);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `operation` <Property {...properties.operation} />

##### Explanation

With `mapItems` you can apply an `operation` to each item in an array to transform it.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `mapItems` can be used.

##### Mark duplicates

```ts
const MarkedArraySchema = v.pipe(
  v.array(v.string()),
  v.mapItems((item, index, array) => {
    const isDuplicate = array.indexOf(item) !== index;
    return { item, isDuplicate };
  })
);
```

#### Related

The following APIs can be combined with `mapItems`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxBytes

Creates a max [bytes](https://en.wikipedia.org/wiki/Byte) validation action.

```ts
const Action = v.maxBytes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxBytes` you can validate the bytes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxBytes` can be used.

##### Max bytes schema

Schema to validate a string with a maximum of 64 bytes.

```ts
const MaxBytesSchema = v.pipe(
  v.string(),
  v.maxBytes(64, 'The string must not exceed 64 bytes.')
);
```

#### Related

The following APIs can be combined with `maxBytes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxEntries

Creates a max entries validation action.

```ts
const Action = v.maxEntries<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxEntries` you can validate the number of entries of an object. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxEntries` can be used.

##### Maximum object entries

Schema to validate an object with a maximum of 5 entries.

```ts
const MaxEntriesSchema = v.pipe(
  v.record(v.string(), v.number()),
  v.maxEntries(5, 'Object must not exceed 5 entries.')
);
```

#### Related

The following APIs can be combined with `maxEntries`.

##### Schemas

<ApiList
  items={[
    'looseObject',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxGraphemes

Creates a max [graphemes](https://en.wikipedia.org/wiki/Grapheme) validation action.

```ts
const Action = v.maxGraphemes<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxGraphemes` you can validate the graphemes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

> Hint: The number of characters per grapheme is not limited. You may want to consider combining `maxGraphemes` with <Link href="../maxLength/">`maxLength`</Link> or <Link href="../maxBytes/">`maxBytes`</Link> to set a stricter limit.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxGraphemes` can be used.

##### Max graphemes schema

Schema to validate a string with a maximum of 8 graphemes.

```ts
const MaxGraphemesSchema = v.pipe(
  v.string(),
  v.maxGraphemes(8, 'The string must not exceed 8 graphemes.')
);
```

#### Related

The following APIs can be combined with `maxGraphemes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxLength

Creates a max length validation action.

```ts
const Action = v.maxLength<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxLength` you can validate the length of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxLength` can be used.

##### Maximum string length

Schema to validate a string with a maximum length of 32 characters.

```ts
const MaxStringSchema = v.pipe(
  v.string(),
  v.maxLength(32, 'The string must not exceed 32 characters.')
);
```

##### Maximum array length

Schema to validate an array with a maximum length of 5 items.

```ts
const MaxArraySchema = v.pipe(
  v.array(v.number()),
  v.maxLength(5, 'The array must not exceed 5 numbers.')
);
```

#### Related

The following APIs can be combined with `maxLength`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxSize

Creates a max size validation action.

```ts
const Action = v.maxSize<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxSize` you can validate the size of a map, set or blob. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxSize` can be used.

##### Blob size schema

Schema to validate a blob with a maximum size of 10 MB.

```ts
const BlobSchema = v.pipe(
  v.blob(),
  v.maxSize(10 * 1024 * 1024, 'The blob must not exceed 10 MB.')
);
```

##### Set size schema

Schema to validate a set with a maximum of 8 numbers.

```ts
const SetSchema = v.pipe(
  v.set(number()),
  v.maxSize(8, 'The set must not exceed 8 numbers.')
);
```

#### Related

The following APIs can be combined with `maxSize`.

##### Schemas

<ApiList
  items={['any', 'blob', 'custom', 'file', 'instance', 'map', 'set', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxValue

Creates a max value validation action.

```ts
const Action = v.maxValue<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxValue` you can validate the value of a string, number, boolean or date. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxValue` can be used.

##### Number schema

Schema to validate a number with a maximum value.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.maxValue(100, 'The number must not exceed 100.')
);
```

##### Date schema

Schema to validate a date with a maximum year.

```ts
const DateSchema = v.pipe(
  v.date(),
  v.maxValue(new Date('1999-12-31'), 'The date must not exceed the year 1999.')
);
```

#### Related

The following APIs can be combined with `maxValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### maxWords

Creates a max [words](https://en.wikipedia.org/wiki/Word) validation action.

```ts
const Action = v.maxWords<TInput, TLocales, TRequirement, TMessage>(
  locales,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `locales` <Property {...properties.locales} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `maxWords` you can validate the words of a string based on the specified `locales`. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `maxWords` can be used.

##### Max words schema

Schema to validate a string with a maximum of 300 words.

```ts
const MaxWordsSchema = v.pipe(
  v.string(),
  v.maxWords('en', 300, 'The string must not exceed 300 words.')
);
```

#### Related

The following APIs can be combined with `maxWords`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### metadata

Creates a custom metadata action.

```ts
const Action = v.metadata<TInput, TMetadata>(metadata_);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMetadata` <Property {...properties.TMetadata} />

#### Parameters

- `metadata_` <Property {...properties['metadata_']} />

##### Explanation

With `metadata` you can attach custom metadata to a schema. This can be useful when working with AI tools or for documentation purposes.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `metadata` can be used.

##### Profile table schema

Schema to describe a profile table.

```ts
const ProfileTableSchema = v.pipe(
  v.object({
    username: v.pipe(v.string(), v.nonEmpty()),
    email: v.pipe(v.string(), v.email()),
    avatar: v.pipe(v.string(), v.url()),
    description: v.pipe(v.string(), v.maxLength(500)),
  }),
  v.metadata({
    table: 'profiles',
    primaryKey: 'username',
    indexes: ['email'],
  })
);
```

#### Related

The following APIs can be combined with `metadata`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['getMetadata', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### mimeType

Creates a [MIME type](https://developer.mozilla.org/docs/Web/HTTP/Basics_of_HTTP/MIME_types) validation action.

```ts
const Action = v.mimeType<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `mimeType` you can validate the MIME type of a blob. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `mimeType` can be used.

##### Image schema

Schema to validate an image file.

```ts
const ImageSchema = v.pipe(
  v.blob(),
  v.mimeType(['image/jpeg', 'image/png'], 'Please select a JPEG or PNG file.')
);
```

#### Related

The following APIs can be combined with `mimeType`.

##### Schemas

<ApiList items={['any', 'blob', 'custom', 'file', 'instance', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minBytes

Creates a min [bytes](https://en.wikipedia.org/wiki/Byte) validation action.

```ts
const Action = v.minBytes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minBytes` you can validate the bytes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minBytes` can be used.

##### Min bytes schema

Schema to validate a string with a minimum of 64 bytes.

```ts
const MinBytesSchema = v.pipe(
  v.string(),
  v.minBytes(64, 'The string must contain at least 64 bytes.')
);
```

#### Related

The following APIs can be combined with `minBytes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minEntries

Creates a min entries validation action.

```ts
const Action = v.minEntries<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minEntries` you can validate the number of entries of an object. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minEntries` can be used.

##### Minimum object entries

Schema to validate an object with a minimum of 5 entries.

```ts
const MinEntriesSchema = v.pipe(
  v.record(v.string(), v.number()),
  v.minEntries(5, 'The object should have at least 5 entries.')
);
```

#### Related

The following APIs can be combined with `minEntries`.

##### Schemas

<ApiList
  items={[
    'looseObject',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minGraphemes

Creates a min [graphemes](https://en.wikipedia.org/wiki/Grapheme) validation action.

```ts
const Action = v.minGraphemes<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minGraphemes` you can validate the graphemes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minGraphemes` can be used.

##### Min graphemes schema

Schema to validate a string with a minimum of 8 graphemes.

```ts
const MinGraphemesSchema = v.pipe(
  v.string(),
  v.minGraphemes(8, 'The string must contain at least 8 graphemes.')
);
```

#### Related

The following APIs can be combined with `minGraphemes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minLength

Creates a min length validation action.

```ts
const Action = v.minLength<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minLength` you can validate the length of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minLength` can be used.

##### Minimum string length

Schema to validate a string with a minimum length of 3 characters.

```ts
const MinStringSchema = v.pipe(
  v.string(),
  v.minLength(3, 'The string must be 3 or more characters long.')
);
```

##### Minimum array length

Schema to validate an array with a minimum length of 5 items.

```ts
const MinArraySchema = v.pipe(
  v.array(v.number()),
  v.minLength(5, 'The array must contain 5 numbers or more.')
);
```

#### Related

The following APIs can be combined with `minLength`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minSize

Creates a min size validation action.

```ts
const Action = v.minSize<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minSize` you can validate the size of a map, set or blob. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minSize` can be used.

##### Blob size schema

Schema to validate a blob with a minimum size of 10 MB.

```ts
const BlobSchema = v.pipe(
  v.blob(),
  v.minSize(10 * 1024 * 1024, 'The blob must be at least 10 MB.')
);
```

##### Set size schema

Schema to validate a set with a minimum of 8 numbers.

```ts
const SetSchema = v.pipe(
  v.set(number()),
  v.minSize(8, 'The set must contain at least 8 numbers.')
);
```

#### Related

The following APIs can be combined with `minSize`.

##### Schemas

<ApiList
  items={['any', 'blob', 'custom', 'file', 'instance', 'map', 'set', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minValue

Creates a min value validation action.

```ts
const Action = v.minValue<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minValue` you can validate the value of a string, number, boolean or date. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minValue` can be used.

##### Number schema

Schema to validate a number with a minimum value.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.minValue(100, 'The number must be at least 100.')
);
```

##### Date schema

Schema to validate a date with a minimum year.

```ts
const DateSchema = v.pipe(
  v.date(),
  v.minValue(new Date('2000-01-01'), 'The date must be after the year 1999.')
);
```

#### Related

The following APIs can be combined with `minValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### minWords

Creates a min [words](https://en.wikipedia.org/wiki/Word) validation action.

```ts
const Action = v.minWords<TInput, TLocales, TRequirement, TMessage>(
  locales,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `locales` <Property {...properties.locales} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `minWords` you can validate the words of a string based on the specified `locales`. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `minWords` can be used.

##### Min words schema

Schema to validate a string with a minimum of 50 words.

```ts
const MinWordsSchema = v.pipe(
  v.string(),
  v.minWords('en', 50, 'The string must contain at least 50 words.')
);
```

#### Related

The following APIs can be combined with `minWords`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### multipleOf

Creates a [multiple](<https://en.wikipedia.org/wiki/Multiple_(mathematics)>) of validation action.

```ts
const Action = v.multipleOf<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `multipleOf` you can validate the value of a number. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `multipleOf` can be used.

##### Even number schema

Schema to validate an even number.

```ts
const EvenNumberSchema = v.pipe(
  v.number(),
  v.multipleOf(2, 'The number must be even.')
);
```

#### Related

The following APIs can be combined with `multipleOf`.

##### Schemas

<ApiList items={['any', 'bigint', 'custom', 'number', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### nanoid

Creates a [Nano ID](https://github.com/ai/nanoid) validation action.

```ts
const Action = v.nanoid<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `nanoid` you can validate the formatting of a string. If the input is not an Nano ID, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `nanoid` can be used.

> Since Nano IDs are not limited to a fixed length, it is recommended to combine `nanoid` with <Link href="../length/">`length`</Link> to ensure the correct length.

##### Nano ID schema

Schema to validate a Nano ID.

```ts
const NanoIdSchema = v.pipe(
  v.string(),
  v.nanoid('The Nano ID is badly formatted.'),
  v.length(21, 'The Nano ID must be 21 characters long.')
);
```

#### Related

The following APIs can be combined with `nanoid`.

##### Schemas

<ApiList items={['any', 'string', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### nonEmpty

Creates a non-empty validation action.

```ts
const Action = v.nonEmpty<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `nonEmpty` you can validate that a string or array is non-empty. If the input is empty, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `nonEmpty` can be used.

##### String schema

Schema to validate that a string is non-empty.

```ts
const StringSchema = v.pipe(
  v.string(),
  v.nonEmpty('The string should contain at least one character.')
);
```

##### Array schema

Schema to validate that an array is non-empty.

```ts
const ArraySchema = v.pipe(
  v.array(v.number()),
  v.nonEmpty('The array should contain at least one item.')
);
```

#### Related

The following APIs can be combined with `nonEmpty`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### normalize

Creates a normalize transformation action.

```ts
const Action = v.normalize<TForm>(form);
```

#### Generics

- `TForm` <Property {...properties.TForm} />

#### Parameters

- `form` <Property {...properties.form} />

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `normalize` can be used.

##### Normalized string

Schema to normalize a string.

```ts
const StringSchema = v.pipe(v.string(), v.normalize());
```

#### Related

The following APIs can be combined with `normalize`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notBytes

Creates a not [bytes](https://en.wikipedia.org/wiki/Byte) validation action.

```ts
const Action = v.notBytes<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notBytes` you can validate the bytes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notBytes` can be used.

##### Not bytes schema

Schema to validate a string with more or less than 8 bytes.

```ts
const NotBytesSchema = v.pipe(
  v.string(),
  v.notBytes(8, 'The string must not have 8 bytes.')
);
```

#### Related

The following APIs can be combined with `notBytes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notEntries

Creates a not entries validation action.

```ts
const Action = v.notEntries<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notEntries` you can validate the number of entries of an object. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notEntries` can be used.

##### Not object entries

Schema to validate an object that does not have 5 entries.

```ts
const NotEntriesSchema = v.pipe(
  v.record(v.string(), v.number()),
  v.notEntries(5, 'Object must not have 5 entries')
);
```

#### Related

The following APIs can be combined with `notEntries`.

##### Schemas

<ApiList
  items={[
    'looseObject',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notGraphemes

Creates a not [graphemes](https://en.wikipedia.org/wiki/Grapheme) validation action.

```ts
const Action = v.notGraphemes<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notGraphemes` you can validate the graphemes of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notGraphemes` can be used.

##### Not graphemes schema

Schema to validate a string with more or less than 8 graphemes.

```ts
const NotGraphemesSchema = v.pipe(
  v.string(),
  v.notGraphemes(8, 'The string must not have 8 graphemes.')
);
```

#### Related

The following APIs can be combined with `notGraphemes`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notLength

Creates a not length validation action.

```ts
const Action = v.notLength<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notLength` you can validate the length of a string or array. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notLength` can be used.

##### String schema

Schema to validate the length of a string.

```ts
const StringSchema = v.pipe(
  v.string(),
  v.notLength(8, 'The string must not be 8 characters long.')
);
```

##### Array schema

Schema to validate the length of an array.

```ts
const ArraySchema = v.pipe(
  v.array(number()),
  v.notLength(10, 'The array must not contain 10 numbers.')
);
```

#### Related

The following APIs can be combined with `notLength`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notSize

Creates a not size validation action.

```ts
const Action = v.notSize<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notSize` you can validate the size of a map, set or blob. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notSize` can be used.

##### Blob size schema

Schema to validate a blob with less ore more then 10 MB.

```ts
const BlobSchema = v.pipe(
  v.blob(),
  v.notSize(10 * 1024 * 1024, 'The blob must not be 10 MB in size.')
);
```

##### Set size schema

Schema to validate a set with less ore more then 8 numbers.

```ts
const SetSchema = v.pipe(
  v.set(number()),
  v.notSize(8, 'The set must not contain 8 numbers.')
);
```

#### Related

The following APIs can be combined with `notSize`.

##### Schemas

<ApiList
  items={['any', 'blob', 'custom', 'file', 'instance', 'map', 'set', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notValue

Creates a not value validation action.

```ts
const Action = v.notValue<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notValue` you can validate the value of a string, number, boolean or date. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notValue` can be used.

##### Number schema

Schema to validate a number that is more or less than 100.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.notValue(100, 'The number must not be 100.')
);
```

##### Date schema

Schema to validate a date that is before or after the start of 2000.

```ts
const DateSchema = v.pipe(
  v.date(),
  v.notValue(new Date('2000-01-01'), 'The date must not be the start of 2000.')
);
```

#### Related

The following APIs can be combined with `notValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notValues

Creates a not values validation action.

```ts
const Action = v.notValues<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notValues` you can validate the value of a string, number, boolean or date. If the input matches one of the values in the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notValues` can be used.

##### Number schema

Schema to validate a number that is not 10, 11 or 12.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.notValues([10, 11, 12], 'The number must not be 10, 11 or 12.')
);
```

#### Related

The following APIs can be combined with `notValues`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### notWords

Creates a not [words](https://en.wikipedia.org/wiki/Word) validation action.

```ts
const Action = v.notWords<TInput, TLocales, TRequirement, TMessage>(
  locales,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `locales` <Property {...properties.locales} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `notWords` you can validate the words of a string based on the specified `locales`. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `notWords` can be used.

##### Not words schema

Schema to validate a string with more or less than 5 words.

```ts
const NotWordsSchema = v.pipe(
  v.string(),
  v.notWords('en', 5, 'The string must not have 5 words.')
);
```

#### Related

The following APIs can be combined with `notWords`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### octal

Creates an [octal](https://en.wikipedia.org/wiki/Octal) validation action.

```ts
const Action = v.octal<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `octal` you can validate the formatting of a string. If the input is not an octal, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `octal` can be used.

##### Octal schema

Schema to validate a octal string.

```ts
const OctalSchema = v.pipe(
  v.string(),
  v.octal('The octal is badly formatted.')
);
```

#### Related

The following APIs can be combined with `octal`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### parseJson

Creates a JSON parse transformation action.

```ts
const Action = v.parseJson<TInput, TConfig, TMessage>(config, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TConfig` <Property {...properties.TConfig} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `config` <Property {...properties.config} />
- `message` <Property {...properties.message} />

##### Explanation

With `parseJson` you can parse a JSON string. If the input is not valid JSON, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `parseJson` can be used.

##### Parse and validate JSON

Parse a JSON string and validate the result.

```ts
const StringifiedObjectSchema = v.pipe(
  v.string(),
  v.parseJson(),
  v.object({ key: v.string() })
);
```

##### Parse JSON with reviver

Parse a JSON string with a reviver function.

```ts
const StringifiedObjectSchema = v.pipe(
  v.string(),
  v.parseJson({
    reviver: (key, value) =>
      typeof value === 'string' ? value.toUpperCase() : value,
  }),
  v.object({ key: v.string() })
);
```

#### Related

The following APIs can be combined with `parseJson`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### partialCheck

Creates a partial check validation action.

```ts
const Action = v.partialCheck<TInput, TPaths, TSelection, TMessage>(
  paths,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TPaths` <Property {...properties.TPaths} />
- `TSelection` <Property {...properties.TSelection} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `paths` <Property {...properties.paths} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `partialCheck` you can freely validate the selected input and return `true` if it is valid or `false` otherwise. If the input does not match your `requirement`, you can use `message` to customize the error message.

> The difference to <Link href='../check/'>`check`</Link> is that `partialCheck` can be executed whenever the selected part of the data is valid, while <Link href='../check/'>`check`</Link> is executed only when the entire dataset is typed. This can be an important advantage when working with forms.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `partialCheck` can be used.

##### Register schema

Schema that ensures that the two passwords match.

```ts
const RegisterSchema = v.pipe(
  v.object({
    email: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your email.'),
      v.email('The email address is badly formatted.')
    ),
    password1: v.pipe(
      v.string(),
      v.nonEmpty('Please enter your password.'),
      v.minLength(8, 'Your password must have 8 characters or more.')
    ),
    password2: v.string(),
  }),
  v.forward(
    v.partialCheck(
      [['password1'], ['password2']],
      (input) => input.password1 === input.password2,
      'The two passwords do not match.'
    ),
    ['password2']
  )
);
```

#### Related

The following APIs can be combined with `partialCheck`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'custom',
    'instance',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'union',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['forward', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### rawCheck

Creates a raw check validation action.

```ts
const Action = v.rawCheck<TInput>(action);
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Parameters

- `action` <Property {...properties.action} />

##### Explanation

With `rawCheck` you can freely validate the input with a custom `action` and add issues if necessary.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `rawCheck` can be used.

##### Emails schema

Object schema that ensures that the primary email is not the same as any of the other emails.

> This `rawCheck` validation action adds an issue for any invalid other email and forwards it via `path` to the appropriate nested field.

```ts
const EmailsSchema = v.pipe(
  v.object({
    primaryEmail: v.pipe(v.string(), v.email()),
    otherEmails: v.array(v.pipe(v.string(), v.email())),
  }),
  v.rawCheck(({ dataset, addIssue }) => {
    if (dataset.typed) {
      dataset.value.otherEmails.forEach((otherEmail, index) => {
        if (otherEmail === dataset.value.primaryEmail) {
          addIssue({
            message: 'This email is already being used as the primary email.',
            path: [
              {
                type: 'object',
                origin: 'value',
                input: dataset.value,
                key: 'otherEmails',
                value: dataset.value.otherEmails,
              },
              {
                type: 'array',
                origin: 'value',
                input: dataset.value.otherEmails,
                key: index,
                value: otherEmail,
              },
            ],
          });
        }
      });
    }
  })
);
```

#### Related

The following APIs can be combined with `rawCheck`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['forward', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### rawTransform

Creates a raw transformation action.

```ts
const Action = v.rawTransform<TInput, TOutput>(action);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `action` <Property {...properties.action} />

##### Explanation

With `rawTransform` you can freely transform and validate the input with a custom `action` and add issues if necessary.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `rawTransform` can be used.

##### Calculate game result

Schema that calculates the total score of a game based on the scores and a multiplier.

> This `rawTransform` validation action adds an issue for points that exceed a certain maximum and forwards it via `path` to the appropriate nested score.

```ts
const GameResultSchema = v.pipe(
  v.object({
    scores: v.array(v.pipe(v.number(), v.integer())),
    multiplier: v.number(),
  }),
  v.rawTransform(({ dataset, addIssue, NEVER }) => {
    // Create total variable
    let total = 0;

    // Iterate over scores and check points
    for (let index = 0; index < dataset.value.scores.length; index++) {
      // Calculate points by multiplying score with multiplier
      const score = dataset.value.scores[index];
      const points = score * dataset.value.multiplier;

      // Add issue if points exceed maximum of 1,000 points
      if (points > 1_000) {
        addIssue({
          message:
            'The score exceeds the maximum allowed value of 1,000 points.',
          path: [
            {
              type: 'object',
              origin: 'value',
              input: dataset.value,
              key: 'scores',
              value: dataset.value.scores,
            },
            {
              type: 'array',
              origin: 'value',
              input: dataset.value.scores,
              key: index,
              value: score,
            },
          ],
        });

        // Abort transformation
        return NEVER;
      }

      // Add points to total
      total += points;
    }

    // Add calculated total to dataset
    return { ...dataset.value, total };
  })
);
```

#### Related

The following APIs can be combined with `rawTransform`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['forward', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### readonly

Creates a readonly transformation action.

```ts
const Action = v.readonly<TInput>();
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `readonly` can be used.

##### Readonly array

Schema for a readonly array of numbers.

```ts
const ArraySchema = v.pipe(v.array(v.number()), v.readonly());
```

##### Readonly entry

Object schema with an entry marked as readonly.

```ts
const ObjectSchema = v.object({
  name: v.string(),
  username: v.pipe(v.string(), v.readonly()),
  age: v.number(),
});
```

#### Related

The following APIs can be combined with `readonly`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### reduceItems

Creates a reduce items transformation action.

```ts
const Action = v.reduceItems<TInput, TOutput>(operation, initial);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `operation` <Property {...properties.operation} />
- `initial` <Property {...properties.initial} />

##### Explanation

With `reduceItems` you can apply an `operation` to each item in an array to reduce it to a single value.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `reduceItems` can be used.

##### Sum all numbers

Schema that sums all the numbers in an array.

```ts
const SumArraySchema = v.pipe(
  v.array(v.number()),
  v.reduceItems((sum, item) => sum + item, 0)
);
```

#### Related

The following APIs can be combined with `reduceItems`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### regex

Creates a [regex](https://en.wikipedia.org/wiki/Regular_expression) validation action.

```ts
const Action = v.regex<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `regex` you can validate the formatting of a string. If the input does not match the `requirement`, you can use `message` to customize the error message.

> Hint: Be careful with the global flag `g` in your regex pattern, as it can lead to unexpected results. See [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp/test#using_test_on_a_regex_with_the_global_flag) for more information.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `regex` can be used.

##### Pixel string schema

Schema to validate a pixel string.

```ts
const PixelStringSchema = v.pipe(
  v.string(),
  v.regex(/^\d+px$/, 'The pixel string is badly formatted.')
);
```

#### Related

The following APIs can be combined with `regex`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### returns

Creates a function return transformation action.

```ts
const Action = v.returns<TInput, TSchema>(schema);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

##### Explanation

With `returns` you can force the returned value of a function to match the given `schema`.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `returns` can be used.

##### Function schema

Schema of a function that transforms a string to a number.

```ts
const FunctionSchema = v.pipe(
  v.function(),
  v.args(v.tuple([v.pipe(v.string(), v.decimal())])),
  v.returns(v.number())
);
```

#### Related

The following APIs can be combined with `returns`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### rfcEmail

Creates a [RFC email](https://datatracker.ietf.org/doc/html/rfc5322#section-3.4.1) validation action.

```ts
const Action = v.rfcEmail<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `rfcEmail` you can validate the formatting of a string. If the input is not an email, you can use `message` to customize the error message.

> This validation action intentionally validates the entire RFC 5322 specification. If you are interested in an action that only covers common email addresses, please use the <Link href="../email/">`email`</Link> action instead.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `rfcEmail` can be used.

##### Email schema

Schema to validate an email.

```ts
const EmailSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please enter your email.'),
  v.rfcEmail('The email is badly formatted.'),
  v.maxLength(30, 'Your email is too long.')
);
```

#### Related

The following APIs can be combined with `rfcEmail`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### safeInteger

Creates a safe integer validation action.

```ts
const Action = v.safeInteger<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `safeInteger` you can validate the value of a number. If the input is not a safe integer, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `safeInteger` can be used.

##### Safe integer schema

Schema to validate an safe integer.

```ts
const SafeIntegerSchema = v.pipe(
  v.number(),
  v.safeInteger('The number must be a safe integer.')
);
```

#### Related

The following APIs can be combined with `safeInteger`.

##### Schemas

<ApiList items={['any', 'custom', 'number', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### size

Creates a size validation action.

```ts
const Action = v.size<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `size` you can validate the size of a map, set or blob. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `size` can be used.

##### Blob size schema

Schema to validate a blob with a size of 256 bytes.

```ts
const BlobSchema = v.pipe(
  v.blob(),
  v.size(256, 'The blob must be 256 bytes in size.')
);
```

##### Set size schema

Schema to validate a set of 8 numbers.

```ts
const SetSchema = v.pipe(
  v.set(number()),
  v.size(8, 'The set must contain 8 numbers.')
);
```

#### Related

The following APIs can be combined with `size`.

##### Schemas

<ApiList
  items={['any', 'array', 'custom', 'instance', 'string', 'tuple', 'unknown']}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### slug

Creates an [slug](https://en.wikipedia.org/wiki/Clean_URL#Slug) validation action.

```ts
const Action = v.slug<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `slug` you can validate the formatting of a string. If the input is not a URL slug, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `slug` can be used.

##### Slug schema

Schema to validate a slug.

```ts
const SlugSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please provide a slug.'),
  v.slug('The slug is badly formatted.'),
  v.maxLength(100, 'Your slug is too long.')
);
```

#### Related

The following APIs can be combined with `slug`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### someItem

Creates a some item validation action.

```ts
const Action = v.someItem<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `someItem` you can freely validate the items of an array and return `true` if they are valid or `false` otherwise. If not some item matches your `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `someItem` can be used.

##### Unsorted array schema

Schema to validate that an array is not sorted.

```ts
const UnsortedArraySchema = v.pipe(
  v.array(v.number()),
  v.someItem(
    (item, index, array) => array.length === 1 || item < array[index - 1],
    'The numbers must not be sorted in ascending order.'
  )
);
```

#### Related

The following APIs can be combined with `someItem`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### sortItems

Creates a sort items transformation action.

```ts
const Action = v.sortItems<TInput>(operation);
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Parameters

- `operation` <Property {...properties.operation} />

##### Explanation

With `sortItems` you can sort the items of an array based on a custom `operation`. This is a function that takes two items and returns a number. If the number is less than 0, the first item is sorted before the second item. If the number is greater than 0, the second item is sorted before the first. If the number is 0, the order of the items is not changed.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `sortItems` can be used.

##### Sort numbers

Schema that sorts the numbers in an array in ascending order.

```ts
const SortedArraySchema = v.pipe(v.array(v.number()), v.sortItems());
```

#### Related

The following APIs can be combined with `sortItems`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### startsWith

Creates a starts with validation action.

```ts
const Action = v.startsWith<TInput, TRequirement, TMessage>(
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `startsWith` you can validate the start of a string. If the start does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `startsWith` can be used.

##### HTTPS URL schema

Schema to validate a HTTPS URL.

```ts
const HttpsUrlSchema = v.pipe(v.string(), v.url(), v.startsWith('https://'));
```

#### Related

The following APIs can be combined with `startsWith`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### stringifyJson

Creates a JSON stringify transformation action.

```ts
const Action = v.stringifyJson<TInput, TConfig, TMessage>(config, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TConfig` <Property {...properties.TConfig} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `config` <Property {...properties.config} />
- `message` <Property {...properties.message} />

##### Explanation

With `stringifyJson` you can stringify a JSON object. If the input is unable to be stringified, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `stringifyJson` can be used.

##### Stringify JSON

Stringify a JSON object.

```ts
const StringifiedObjectSchema = v.pipe(
  v.object({ key: v.string() }),
  v.stringifyJson()
);
```

##### Stringify JSON with replacer

Stringify a JSON object with a replacer function.

```ts
const StringifiedObjectSchema = v.pipe(
  v.object({ key: v.string() }),
  v.stringifyJson({
    replacer: (key, value) =>
      typeof value === 'string' ? value.toUpperCase() : value,
  })
);
```

#### Related

The following APIs can be combined with `stringifyJson`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'picklist',
    'record',
    'strictObject',
    'strictTuple',
    'string',
    'tuple',
    'tupleWithRest',
    'union',
    'unknown',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### title

Creates a title metadata action.

```ts
const Action = v.title<TInput, TTitle>(title_);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TTitle` <Property {...properties.TTitle} />

#### Parameters

- `title_` <Property {...properties['title_']} />

##### Explanation

With `title` you can give a title to a schema. This can be useful when working with AI tools or for documentation purposes.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `title` can be used.

##### Username schema

Schema to validate a user name.

```ts
const UsernameSchema = v.pipe(
  v.string(),
  v.regex(/^[a-z0-9_-]{4,16}$/iu),
  v.title('Username'),
  v.description(
    'A username must be between 4 and 16 characters long and can only contain letters, numbers, underscores and hyphens.'
  )
);
```

#### Related

The following APIs can be combined with `title`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['getTitle', 'pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### toLowerCase

Creates a to lower case transformation action.

```ts
const Action = v.toLowerCase();
```

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `toLowerCase` can be used.

##### Lower case string

Schema that transforms a string to lower case.

```ts
const StringSchema = v.pipe(v.string(), v.toLowerCase());
```

#### Related

The following APIs can be combined with `toLowerCase`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### toMaxValue

Creates a to max value transformation action.

```ts
const Action = v.toMaxValue<TInput, TRequirement>(requirement);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Parameters

- `requirement` <Property {...properties.requirement} />

##### Explanation

With `toMaxValue` you can enforce a maximum value for a number, date or string. If the input does not meet the `requirement`, it will be changed to its value.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `toMaxValue` can be used.

##### Number schema

Schema to enforce a maximum value for a number.

```ts
const NumberSchema = v.pipe(v.number(), v.toMaxValue(100));
```

##### Date schema

Schema to enforce a maximum value for a date.

```ts
const DateSchema = v.pipe(v.date(), v.toMaxValue(new Date('1999-12-31')));
```

#### Related

The following APIs can be combined with `toMaxValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### toMinValue

Creates a to min value transformation action.

```ts
const Action = v.toMinValue<TInput, TRequirement>(requirement);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Parameters

- `requirement` <Property {...properties.requirement} />

##### Explanation

With `toMinValue` you can enforce a minimum value for a number, date or string. If the input does not meet the `requirement`, it will be changed to its value.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `toMinValue` can be used.

##### Number schema

Schema to enforce a minimum value for a number.

```ts
const NumberSchema = v.pipe(v.number(), v.toMinValue(100));
```

##### Date schema

Schema to enforce a minimum value for a date.

```ts
const DateSchema = v.pipe(v.date(), v.toMinValue(new Date('1999-12-31')));
```

#### Related

The following APIs can be combined with `toMinValue`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### toUpperCase

Creates a to upper case transformation action.

```ts
const Action = v.toUpperCase();
```

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `toUpperCase` can be used.

##### Lower case string

Schema that transforms a string to upper case.

```ts
const StringSchema = v.pipe(v.string(), v.toUpperCase());
```

#### Related

The following APIs can be combined with `toUpperCase`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### transform

Creates a custom transformation action.

```ts
const Action = v.transform<TInput, TOutput>(action);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `action` <Property {...properties.action} />

##### Explanation

`transform` can be used to freely transform the input. The `action` parameter is a function that takes the input and returns the transformed output.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `transform` can be used.

##### Transform to length

Schema that transforms a string to its length.

```ts
const StringLengthSchema = v.pipe(
  v.string(),
  v.transform((input) => input.length)
);
```

##### Add object entry

Schema that transforms an object to add an entry.

```ts
const UserSchema = v.pipe(
  v.object({ name: v.string(), age: v.number() }),
  v.transform((input) => ({
    ...input,
    created: new Date().toISOString(),
  }))
);
```

#### Related

The following APIs can be combined with `transform`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### trim

Creates a trim transformation action.

```ts
const Action = v.trim();
```

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `trim` can be used.

##### Trimmed string

Schema to trim the start and end of a string.

```ts
const StringSchema = v.pipe(v.string(), v.trim());
```

#### Related

The following APIs can be combined with `trim`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### trimEnd

Creates a trim end transformation action.

```ts
const Action = v.trimEnd();
```

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `trimEnd` can be used.

##### Trimmed string

Schema to trimEnd the end of a string.

```ts
const StringSchema = v.pipe(v.string(), v.trimEnd());
```

#### Related

The following APIs can be combined with `trimEnd`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### trimStart

Creates a trim start transformation action.

```ts
const Action = v.trimStart();
```

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `trimStart` can be used.

##### Trimmed string

Schema to trimStart the start of a string.

```ts
const StringSchema = v.pipe(v.string(), v.trimStart());
```

#### Related

The following APIs can be combined with `trimStart`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### ulid

Creates an [ULID](https://github.com/ulid/spec) validation action.

```ts
const Action = v.ulid<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `ulid` you can validate the formatting of a string. If the input is not an ULID, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `ulid` can be used.

##### ULID schema

Schema to validate an ULID.

```ts
const UlidSchema = v.pipe(v.string(), v.ulid('The ULID is badly formatted.'));
```

#### Related

The following APIs can be combined with `ulid`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### url

Creates an [URL](https://en.wikipedia.org/wiki/URL) validation action.

```ts
const Action = v.url<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `url` you can validate the formatting of a string. If the input is not an URL, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `url` can be used.

##### URL schema

Schema to validate an URL.

```ts
const UrlSchema = v.pipe(
  v.string(),
  v.nonEmpty('Please enter your url.'),
  v.url('The url is badly formatted.'),
  v.endsWith('.com', 'Only ".com" domains are allowed.')
);
```

#### Related

The following APIs can be combined with `url`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### uuid

Creates an [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) validation action.

```ts
const Action = v.uuid<TInput, TMessage>(message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `message` <Property {...properties.message} />

##### Explanation

With `uuid` you can validate the formatting of a string. If the input is not an UUID, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `uuid` can be used.

##### UUID schema

Schema to validate an UUID.

```ts
const UuidSchema = v.pipe(v.string(), v.uuid('The UUID is badly formatted.'));
```

#### Related

The following APIs can be combined with `uuid`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### value

Creates a value validation action.

```ts
const Action = v.value<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `value` you can validate the value of a string, number, boolean or date. If the input does not match the `requirement`, you can use `message` to customize the error message.

> This action does not change the type of the pipeline. Use the <Link href="../literal/">`literal`</Link> schema instead if you want the type to match a specific value.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `value` can be used.

##### Number schema

Schema to validate a number with a specific value.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.value(100, 'The number must be 100.')
);
```

##### Date schema

Schema to validate a date with a specific value.

```ts
const DateSchema = v.pipe(
  v.date(),
  v.value(new Date('2000-01-01'), 'The date must be the first day of 2000.')
);
```

#### Related

The following APIs can be combined with `value`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### values

Creates a values validation action.

```ts
const Action = v.values<TInput, TRequirement, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `values` you can validate the value of a string, number, boolean or date. If the input does not match one of the values in the `requirement`, you can use `message` to customize the error message.

> This action does not change the type of the pipeline. Use the <Link href="../picklist/">`picklist`</Link> schema instead if you want the type to match the union of specific values.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `values` can be used.

##### Number schema

Schema to validate a number with specific values.

```ts
const NumberSchema = v.pipe(
  v.number(),
  v.values([5, 15, 20], 'The number must be one of the allowed numbers.')
);
```

#### Related

The following APIs can be combined with `values`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'boolean',
    'custom',
    'date',
    'number',
    'string',
    'unknown',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

### words

Creates a [words](https://en.wikipedia.org/wiki/Word) validation action.

```ts
const Action = v.words<TInput, TLocales, TRequirement, TMessage>(
  locales,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `locales` <Property {...properties.locales} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `words` you can validate the words of a string based on the specified `locales`. If the input does not match the `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `words` can be used.

##### Words schema

Schema to validate a string with 3 words.

```ts
const WordsSchema = v.pipe(
  v.string(),
  v.words('en', 3, 'Exactly 3 words are required.')
);
```

#### Related

The following APIs can be combined with `words`.

##### Schemas

<ApiList items={['any', 'custom', 'string']} />

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

## Storages (API)

### deleteGlobalConfig

Deletes the global configuration.

```ts
v.deleteGlobalConfig();
```

### deleteGlobalMessage

Deletes a global error message.

```ts
v.deleteGlobalMessage(lang);
```

#### Parameters

- `lang` <Property {...properties.lang} />

### deleteSchemaMessage

Deletes a schema error message.

```ts
v.deleteSchemaMessage(lang);
```

#### Parameters

- `lang` <Property {...properties.lang} />

### deleteSpecificMessage

Deletes a specific error message.

```ts
v.deleteSpecificMessage(reference, lang);
```

#### Parameters

- `reference` <Property {...properties.reference} />
- `lang` <Property {...properties.lang} />

### getGlobalConfig

Returns the global configuration.

```ts
const config = v.getGlobalConfig<TIssue>(merge);
```

#### Generics

- `TIssue` <Property {...properties.TIssue} />

#### Parameters

- `merge` <Property {...properties.merge} />

##### Explanation

Properties that you want to explicitly override can be optionally specified with `merge`.

#### Returns

- `config` <Property {...properties.config} />

### getGlobalMessage

Returns a global error message.

```ts
const message = v.getGlobalMessage(lang);
```

#### Parameters

- `lang` <Property {...properties.lang} />

#### Returns

- `message` <Property {...properties.message} />

### getSchemaMessage

Returns a schema error message.

```ts
const message = v.getSchemaMessage(lang);
```

#### Parameters

- `lang` <Property {...properties.lang} />

#### Returns

- `message` <Property {...properties.message} />

### getSpecificMessage

Returns a specific error message.

```ts
const message = v.getSpecificMessage(reference, lang);
```

#### Parameters

- `reference` <Property {...properties.reference} />
- `lang` <Property {...properties.lang} />

#### Returns

- `message` <Property {...properties.message} />

### setGlobalConfig

Sets the global configuration.

```ts
v.setGlobalConfig(merge);
```

#### Parameters

- `config` <Property {...properties.config} />

##### Explanation

The properties specified by `config` are merged with the existing global configuration. If a property is already set, it will be overwritten.

### setGlobalMessage

Sets a global error message.

```ts
v.setGlobalMessage(message, lang);
```

#### Parameters

- `message` <Property {...properties.message} />
- `lang` <Property {...properties.lang} />

### setSchemaMessage

Sets a schema error message.

```ts
v.setSchemaMessage(message, lang);
```

#### Parameters

- `message` <Property {...properties.message} />
- `lang` <Property {...properties.lang} />

### setSpecificMessage

Sets a specific error message.

```ts
v.setSpecificMessage<TReference>(reference, message, lang);
```

#### Generics

- `TReference` <Property {...properties.TReference} />

#### Parameters

- `reference` <Property {...properties.reference} />
- `message` <Property {...properties.message} />
- `lang` <Property {...properties.lang} />

## Utils (API)

### entriesFromList

Creates an object entries definition from a list of keys and a schema.

```ts
const entries = v.entriesFromList<TList, TSchema>(list, schema);
```

#### Generics

- `TList` <Property {...properties.TList} />
- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `list` <Property {...properties.list} />
- `schema` <Property {...properties.schema} />

#### Returns

- `entries` <Property {...properties.entries} />

#### Examples

The following example show how `entriesFromList` can be used.

```ts
const ObjectSchema = v.object(
  v.entriesFromList(['foo', 'bar', 'baz'], v.string())
);
```

#### Related

The following APIs can be combined with `entriesFromList`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'undefined',
    'undefinedable',
    'union',
    'unionWithRest',
    'unknown',
    'variant',
    'void',
  ]}
/>

### entriesFromObjects

Creates a new object entries definition from existing object schemas.

```ts
const entries = v.entriesFromObjects<TSchemas>(schemas);
```

#### Generics

- `TSchemas` <Property {...properties.TSchemas} />

#### Parameters

- `schemas` <Property {...properties.schemas} />

#### Returns

- `entries` <Property {...properties.entries} />

#### Examples

The following example show how `entriesFromObjects` can be used.

> Hint: The third schema of the list overwrites the `foo` and `baz` properties of the previous schemas.

```ts
const ObjectSchema = v.object(
  v.entriesFromObjects([
     v.object({ foo:  v.string(), bar:  v.string() });
     v.object({ baz:  v.number(), qux:  v.number() });
     v.object({ foo:  v.boolean(), baz:  v.boolean() });
  ])
);
```

#### Related

The following APIs can be combined with `entriesFromObjects`.

##### Schemas

<ApiList items={['looseObject', 'object', 'objectWithRest', 'strictObject']} />

### getDotPath

Creates and returns the dot path of an issue if possible.

```ts
const dotPath = v.getDotPath<TSchema>(issue);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `issue` <Property {...properties.issue} />

#### Returns

- `dotPath` <Property {...properties.dotPath} />

### isOfKind

A generic type guard to check the kind of an object.

```ts
const result = v.isOfKind<TKind, TObject>(kind, object);
```

#### Generics

- `TKind` <Property {...properties.TKind} />
- `TObject` <Property {...properties.TObject} />

#### Parameters

- `kind` <Property {...properties.kind} />
- `object` <Property {...properties.object} />

#### Returns

- `result` <Property {...properties.result} />

### isOfType

A generic type guard to check the type of an object.

```ts
const result = v.isOfType<TType, TObject>(type, object);
```

#### Generics

- `TType` <Property {...properties.TType} />
- `TObject` <Property {...properties.TObject} />

#### Parameters

- `type` <Property {...properties.type} />
- `object` <Property {...properties.object} />

#### Returns

- `result` <Property {...properties.result} />

### isValiError

A type guard to check if an error is a ValiError.

```ts
const result = v.isValiError<TSchema>(error);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `error` <Property {...properties.error} />

#### Returns

- `result` <Property {...properties.result} />

### ValiError

Creates a Valibot error with useful information.

```ts
const error = new v.ValiError<TSchema>(issues);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `issues` <Property {...properties.issues} />

#### Returns

- `error` <Property {...properties.error} />

## Async (API)

### argsAsync

Creates a function arguments transformation action.

```ts
const Action = v.argsAsync<TInput, TSchema>(schema);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

##### Explanation

With `argsAsync` you can force the arguments of a function to match the given `schema`.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `argsAsync` can be used.

##### Product function schema

Schema of a function that returns a product by its ID.

```ts
import { isValidProductId } from '~/api';

const ProductFunctionSchema = v.pipeAsync(
  v.function(),
  v.argsAsync(
    v.tupleAsync([v.pipeAsync(v.string(), v.checkAsync(isValidProductId))])
  ),
  v.returnsAsync(
    v.pipeAsync(
      v.promise(),
      v.awaitAsync(),
      v.object({
        id: v.string(),
        name: v.string(),
        price: v.number(),
      })
    )
  )
);
```

#### Related

The following APIs can be combined with `argsAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'custom',
    'looseTuple',
    'function',
    'strictTuple',
    'tuple',
    'tupleWithRest',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'customAsync',
    'looseTupleAsync',
    'pipeAsync',
    'returnsAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
  ]}
/>

### arrayAsync

Creates an array schema.

```ts
const Schema = v.arrayAsync<TItem, TMessage>(item, message);
```

#### Generics

- `TItem` <Property {...properties.TItem} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `item` <Property {...properties.item} />
- `message` <Property {...properties.message} />

##### Explanation

With `arrayAsync` you can validate the data type of the input. If the input is not an array, you can use `message` to customize the error message.

> If your array has a fixed length, consider using <Link href="../tupleAsync/">`tupleAsync`</Link> for a more precise typing.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `arrayAsync` can be used.

##### Stored emails schema

Schema to validate an array of stored emails.

```ts
import { isEmailPresent } from '~/api';

const StoredEmailsSchema = v.arrayAsync(
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  )
);
```

#### Related

The following APIs can be combined with `arrayAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### awaitAsync

Creates an await transformation action.

```ts
const Action = v.awaitAsync<TInput>();
```

#### Generics

- `TInput` <Property {...properties.TInput} />

##### Explanation

With `awaitAsync` you can transform a promise into its resolved value.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `awaitAsync` can be used.

##### Unique emails schema

Schema to check a set of emails wrapped in a promise object.

```ts
const UniqueEmailsSchema = v.pipeAsync(
  v.promise(),
  v.awaitAsync(),
  v.set(v.pipe(v.string(), v.email()))
);
```

#### Related

The following APIs can be combined with `awaitAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### checkAsync

Creates a check validation action.

```ts
const Action = v.checkAsync<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `checkAsync` you can freely validate the input and return `true` if it is valid or `false` otherwise. If the input does not match your `requirement`, you can use `message` to customize the error message.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `checkAsync` can be used.

##### Cart item schema

Schema to check a cart item object.

```ts
import { getProductItem } from '~/api';

const CartItemSchema = v.pipeAsync(
  v.object({
    itemId: v.pipe(v.string(), v.regex(/^[a-z0-9]{10}$/i)),
    quantity: v.pipe(v.number(), v.minValue(1)),
  }),
  v.checkAsync(async (input) => {
    const productItem = await getProductItem(input.itemId);
    return productItem?.quantity >= input.quantity;
  }, 'The required quantity is greater than available.')
);
```

#### Related

The following APIs can be combined with `checkAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'customAsync',
    'exactOptionalAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### checkItemsAsync

Creates a check items validation action.

```ts
const Action = v.checkItemsAsync<TInput, TMessage>(requirement, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `checkItemsAsync` you can freely validate the items of an array and return `true` if they are valid or `false` otherwise. If an item does not match your `requirement`, you can use `message` to customize the error message.

> The special thing about `checkItemsAsync` is that it automatically forwards each issue to the appropriate item.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `checkItemsAsync` can be used.

##### Cart items schema

Schema to check an array of cart item objects.

```ts
import { getProductItem } from '~/api';

const CartItemsSchema = v.pipeAsync(
  v.array(
    v.object({
      itemId: v.pipe(v.string(), v.uuid()),
      quantity: v.pipe(v.number(), v.minValue(1)),
    })
  ),
  v.checkItemsAsync(async (input) => {
    const productItem = await getProductItem(input.itemId);
    return (productItem?.quantity ?? 0) >= input.quantity;
  }, 'The required quantity is greater than available.')
);
```

#### Related

The following APIs can be combined with `checkItemsAsync`.

##### Schemas

<ApiList items={['any', 'array', 'custom', 'instance', 'tuple', 'unknown']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList items={['arrayAsync', 'customAsync', 'pipeAsync', 'tupleAsync']} />

### customAsync

Creates a custom schema.

> This schema function allows you to define a schema that matches a value based on a custom function. Use it whenever you need to define a schema that cannot be expressed using any of the other schema functions.

```ts
const Schema = v.customAsync<TInput, TMessage>(check, message);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `check` <Property {...properties.check} />
- `message` <Property {...properties.message} />

##### Explanation

With `customAsync` you can validate the data type of the input. If the input does not match the validation of `check`, you can use `message` to customize the error message.

> Make sure that the validation in `check` matches the data type of `TInput`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `customAsync` can be used.

##### Vacant seat schema

Schema to validate a vacant seat.

```ts
import { isSeatVacant } from '~/api';

type Group = 'A' | 'B' | 'C' | 'D' | 'E';
type DigitLessThanSix = '0' | '1' | '2' | '3' | '4' | '5';
type Digit = DigitLessThanSix | '6' | '7' | '8' | '9';
type Seat = `${Group}${DigitLessThanSix}${Digit}`;

function isSeat(possibleSeat: string): possibleSeat is Seat {
  return /^[A-E][0-5]\d$/.test(possibleSeat);
}

const VacantSeatSchema = v.customAsync<Seat>(
  (input) => typeof input === 'string' && isSeat(input) && isSeatVacant(input),
  'The input is not a valid vacant seat.'
);
```

#### Related

The following APIs can be combined with `customAsync`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
  ]}
/>

### exactOptionalAsync

Creates an exact optional schema.

```ts
const Schema = v.exactOptionalAsync<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `exactOptionalAsync` the validation of your schema will pass missing object entries, and if you specify a `default_` input value, the schema will use it if the object entry is missing. For this reason, the output type may differ from the input type of the schema.

> The difference to <Link href="../optionalAsync/">`optionalAsync`</Link> is that this schema function follows the implementation of TypeScript's [`exactOptionalPropertyTypes` configuration](https://www.typescriptlang.org/tsconfig/#exactOptionalPropertyTypes) and only allows missing but not undefined object entries.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `exactOptionalAsync` can be used.

##### New user schema

Schema to validate new user details.

```ts
import { isEmailUnique, isUsernameUnique } from '~/api';

const NewUserSchema = v.objectAsync({
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailUnique, 'The email is not unique.')
  ),
  username: v.exactOptionalAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

/*
  The input and output types of the schema:
    {
      email: string;
      password: string;
      username?: string;
    }
*/
```

##### Unwrap exact optional schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `exactOptionalAsync`.

```ts
import { isUsernameUnique } from '~/api';

const UsernameSchema = v.unwrap(
  // Assume this schema is from a different file and is reused here
  v.exactOptionalAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `exactOptionalAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### fallbackAsync

Returns a fallback value as output if the input does not match the schema.

```ts
const Schema = v.fallbackAsync<TSchema, TFallback>(schema, fallback);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TFallback` <Property {...properties.TFallback} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `fallback` <Property {...properties.fallback} />

##### Explanation

`fallbackAsync` allows you to define a fallback value for the output that will be used if the validation of the input fails. This means that no issues will be returned when using `fallbackAsync` and the schema will always return an output.

> If you only want to set a default value for `null` or `undefined` inputs, you should use <Link href="../optionalAsync/">`optionalAsync`</Link>, <Link href="../nullableAsync/">`nullableAsync`</Link> or <Link href="../nullishAsync/">`nullishAsync`</Link> instead.

> The fallback value is not validated. Make sure that the fallback value matches your schema.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `fallbackAsync` can be used.

##### Unique username schema

Schema that will always return a unique username.

> By using a function as the `fallbackAsync` parameter, the schema will return any unique username each time the input does not match the schema.

```ts
import { getAnyUniqueUsername, isUsernameUnique } from '~/api';

const UniqueUsernameSchema = v.fallbackAsync(
  v.pipeAsync(v.string(), v.minLength(4), v.checkAsync(isUsernameUnique)),
  getAnyUniqueUsername
);
```

#### Related

The following APIs can be combined with `fallbackAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'config',
    'getDefault',
    'getFallback',
    'keyof',
    'message',
    'omit',
    'pick',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### forwardAsync

Forwards the issues of the passed validation action.

```ts
const Action = v.forwardAsync<TInput, TIssue, TPath>(action, path);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TIssue` <Property {...properties.TIssue} />
- `TPath` <Property {...properties.TPath} />

#### Parameters

- `action` <Property {...properties.action} />
- `path` <Property {...properties.path} />

##### Explanation

`forwardAsync` allows you to forward the issues of the passed validation `action` via `path` to a nested field of a schema.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `forwardAsync` can be used.

##### Allowed action schema

Schema that checks if the user is allowed to complete an action.

```ts
import { isAllowedAction, isUsernamePresent } from '~/api';

const AllowedActionSchema = v.pipeAsync(
  v.objectAsync({
    username: v.pipeAsync(
      v.string(),
      v.minLength(3),
      v.checkAsync(isUsernamePresent, 'The username is not in the database.')
    ),
    action: v.picklist(['view', 'edit', 'delete']),
  }),
  v.forwardAsync(
    v.checkAsync(
      isAllowedAction,
      'The user is not allowed to complete the action.'
    ),
    ['action']
  )
);
```

#### Related

The following APIs can be combined with `forwardAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'custom',
    'looseObject',
    'looseTuple',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'union',
    'unknown',
    'variant',
  ]}
/>

##### Methods

<ApiList items={['omit', 'pick']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'reduceItems',
    'someItem',
    'sortItems',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'objectAsync',
    'objectWithRestAsync',
    'partialAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'recordAsync',
    'requiredAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### getDefaultsAsync

Returns the default values of the schema.

> The difference to <Link href='../getDefault/'>`getDefault`</Link> is that for object and tuple schemas this function recursively returns the default values of the subschemas instead of `undefined`.

```ts
const values = v.getDefaultsAsync<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `values` <Property {...properties.values} />

#### Examples

The following examples show how `getDefaultsAsync` can be used.

##### Donation schema defaults

Get the default values of a donation schema.

```ts
import { getRandomOrgId } from '~/api';

const DonationSchema = v.objectAsync({
  timestamp: v.optional(v.date(), () => new Date()),
  sponsor: v.optional(v.pipe(v.string(), v.nonEmpty()), 'anonymous'),
  organizationId: v.optionalAsync(v.pipe(v.string(), v.uuid()), getRandomOrgId),
  message: v.optional(v.pipe(v.string(), v.minLength(1))),
});

const defaultValues = await v.getDefaultsAsync(DonationSchema);

/*
  {
    timestamp: new Date(),
    sponsor: "anonymous",
    organizationId: "43775869-95f3-4e00-9f37-161ec8f9f7cd",
    message: undefined
  }
*/
```

#### Related

The following APIs can be combined with `getDefaultsAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### getFallbacksAsync

Returns the fallback values of the schema.

> The difference to <Link href='../getFallback/'>`getFallback`</Link> is that for object and tuple schemas this function recursively returns the fallback values of the subschemas instead of `undefined`.

```ts
const values = v.getFallbacksAsync<TSchema>(schema);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

#### Returns

- `values` <Property {...properties.values} />

#### Examples

The following examples show how `getFallbacksAsync` can be used.

##### New user fallbacks

Get the fallback values of a new user schema.

```ts
import { getAnyUniqueUsername, isUsernameUnique } from '~/api';

const NewUserSchema = v.objectAsync({
  username: v.fallbackAsync(
    v.pipeAsync(v.string(), v.minLength(3), v.checkAsync(isUsernameUnique)),
    getAnyUniqueUsername
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

const fallbackValues = await v.getFallbacksAsync(NewUserSchema);
/*
  {
    username: "cookieMonster07",
    password: undefined
  }
*/
```

#### Related

The following APIs can be combined with `getFallbacksAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'unwrap',
  ]}
/>

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### intersectAsync

Creates an intersect schema.

> I recommend to read the <Link href="/guides/intersections/">intersections guide</Link> before using this schema function.

```ts
const Schema = v.intersectAsync<TOptions, TMessage>(options, message);
```

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `intersectAsync` you can validate if the input matches each of the given `options`. If the output of the intersection cannot be successfully merged, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `intersectAsync` can be used.

##### Donation schema

Schema that combines objects to validate donation details.

```ts
import { isOrganizationPresent } from '~/api';

const DonationSchema = v.intersectAsync([
  v.objectAsync({
    organizationId: v.pipeAsync(
      v.string(),
      v.uuid(),
      v.checkAsync(
        isOrganizationPresent,
        'The organization is not in the database.'
      )
    ),
  }),
  // Assume the schemas below are from different files and are reused here
  v.object({
    amount: v.pipe(v.number(), v.minValue(100)),
    message: v.pipe(v.string(), v.nonEmpty()),
  }),
  v.object({
    amount: v.pipe(v.number(), v.maxValue(1_000_000)),
    message: v.pipe(v.string(), v.maxLength(500)),
  }),
]);
```

#### Related

The following APIs can be combined with `intersectAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexColor',
    'hexadecimal',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### lazyAsync

Creates a lazy schema.

```ts
const Schema = v.lazyAsync<TWrapped>(getter);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />

#### Parameters

- `getter` <Property {...properties.getter} />

##### Explanation

The `getter` function is called lazily to retrieve the schema. This is necessary to be able to access the input through the first argument of the `getter` function and to avoid a circular dependency for recursive schemas.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `lazyAsync` can be used.

##### Transaction list schema

Recursive schema to validate transactions.

> Due to a TypeScript limitation, the input and output types of recursive schemas cannot be inferred automatically. Therefore, you must explicitly specify these types using <Link href="/api/GenericSchemaAsync/">`GenericSchemaAsync`</Link>.

```ts
import { isTransactionValid } from '~/api';

type Transaction = {
  transactionId: string;
  next: Transaction | null;
};

const TransactionSchema: v.GenericSchemaAsync<Transaction> = v.objectAsync({
  transactionId: v.pipeAsync(
    v.string(),
    v.uuid(),
    v.checkAsync(isTransactionValid, 'The transaction is not valid.')
  ),
  next: v.nullableAsync(v.lazyAsync(() => TransactionSchema)),
});
```

##### Email or username schema

Schema to validate an object containing an email or username.

> In most cases, <Link href="/api/unionAsync/">`unionAsync`</Link> and <Link href="/api/variantAsync/">`variantAsync`</Link> are the better choices for creating such a schema. I recommend using `lazyAsync` only in special cases.

```ts
import { isEmailPresent, isUsernamePresent } from '~/api';

const EmailOrUsernameSchema = v.lazyAsync((input) => {
  if (input && typeof input === 'object' && 'type' in input) {
    switch (input.type) {
      case 'email':
        return v.objectAsync({
          type: v.literal('email'),
          email: v.pipeAsync(
            v.string(),
            v.email(),
            v.checkAsync(
              isEmailPresent,
              'The email is not present in the database.'
            )
          ),
        });
      case 'username':
        return v.objectAsync({
          type: v.literal('username'),
          username: v.pipeAsync(
            v.string(),
            v.nonEmpty(),
            v.checkAsync(
              isUsernamePresent,
              'The username is not present in the database.'
            )
          ),
        });
    }
  }
  return v.never();
});
```

#### Related

The following APIs can be combined with `lazyAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexColor',
    'hexadecimal',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### looseObjectAsync

Creates a loose object schema.

```ts
const Schema = v.looseObjectAsync<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `looseObjectAsync` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object, you can use `message` to customize the error message.

> The difference to <Link href="../objectAsync/">`objectAsync`</Link> is that this schema includes any unknown entries in the output. In addition, this schema filters certain entries from the unknown entries for security reasons.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `looseObjectAsync` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### New user schema

Schema to validate a loose object containing specific new user details.

```ts
import { isEmailPresent } from '~/api';

const NewUserSchema = v.looseObjectAsync({
  firstName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  lastName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is already in use by another user.')
  ),
  password: v.pipe(v.string(), v.minLength(8)),
  avatar: v.optional(v.pipe(v.string(), v.url())),
});
```

#### Related

The following APIs can be combined with `looseObjectAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={['config', 'getDefault', 'getFallback', 'keyof', 'omit', 'pick']}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'flavor',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### looseTupleAsync

Creates a loose tuple schema.

```ts
const Schema = v.looseTupleAsync<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `looseTuplAsynce` you can validate the data type of the input and whether the content matches `items`. If the input is not an array, you can use `message` to customize the error message.

> The difference to <Link href="../tupleAsync/">`tupleAsync`</Link> is that this schema does include unknown items into the output.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `looseTupleAsync` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Number and email tuple

Schema to validate a loose tuple with one number and one stored email address.

```ts
import { isEmailPresent } from '~/api';

const TupleSchema = v.looseTupleAsync([
  v.number(),
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  ),
]);
```

#### Related

The following APIs can be combined with `looseTupleAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### mapAsync

Creates a map schema.

```ts
const Schema = v.mapAsync<TKey, TValue, TMessage>(key, value, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `mapAsync` you can validate the data type of the input and whether the entries match `key` and `value`. If the input is not a map, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `mapAsync` can be used.

##### Shopping items schema

Schema to validate a map with usernames that are allowed to shop as keys and the total items purchased as values.

```ts
import { isUserVerified } from '~/api';

const ShoppingItemsSchema = v.mapAsync(
  v.pipeAsync(
    v.string(),
    v.checkAsync(isUserVerified, 'The username is not allowed to shop.')
  ),
  v.pipe(v.number(), v.minValue(0))
);
```

#### Related

The following APIs can be combined with `mapAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'brand',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### nonNullableAsync

Creates a non nullable schema.

> This schema function can be used to override the behavior of <Link href="../nullableAsync/">`nullableAsync`</Link>.

```ts
const Schema = v.nonNullableAsync<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonNullableAsync` the validation of your schema will not pass `null` inputs. If the input is `null`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonNullableAsync` can be used.

##### Unique username schema

Schema to validate a non-null unique username.

```ts
import { isUsernameUnique } from '~/api';

const UniqueUsernameSchema = v.nonNullableAsync(
  // Assume this schema is from a different file and reused here.
  v.nullableAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `nonNullableAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### nonNullishAsync

Creates a non nullish schema.

> This schema function can be used to override the behavior of <Link href="../nullishAsync/">`nullishAsync`</Link>.

```ts
const Schema = v.nonNullishAsync<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonNullishAsync` the validation of your schema will not pass `null` and `undefined` inputs. If the input is `null` or `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonNullishAsync` can be used.

##### Allowed country schema

Schema to check if a string matches one of the allowed country names.

```ts
import { isAllowedCountry } from '~/api';

const AllowedCountrySchema = v.nonNullishAsync(
  // Assume this schema is from a different file and reused here.
  v.nullishAsync(
    v.pipeAsync(v.string(), v.nonEmpty(), v.checkAsync(isAllowedCountry))
  )
);
```

#### Related

The following APIs can be combined with `nonNullishAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### nonOptionalAsync

Creates a non optional schema.

> This schema function can be used to override the behavior of <Link href="../optionalAsync/">`optionalAsync`</Link>.

```ts
const Schema = v.nonOptionalAsync<TWrapped, TMessage>(wrapped, message);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `message` <Property {...properties.message} />

##### Explanation

With `nonOptionalAsync` the validation of your schema will not pass `undefined` inputs. If the input is `undefined`, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nonOptionalAsync` can be used.

##### Add user schema

Schema to validate an object containing details required to add a user to an existing group.

```ts
import { isGroupPresent } from '~/api';

const AddUserSchema = v.objectAsync({
  groupId: v.nonOptionalAsync(
    // Assume this schema is from a different file and reused here.
    v.optionalAsync(
      v.pipeAsync(
        v.string(),
        v.uuid(),
        v.checkAsync(
          isGroupPresent,
          'The group is not present in the database.'
        )
      )
    )
  ),
  userEmail: v.pipe(v.string(), v.email()),
});
```

#### Related

The following APIs can be combined with `nonOptionalAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### nullableAsync

Creates a nullable schema.

```ts
const Schema = v.nullableAsync<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `nullableAsync` the validation of your schema will pass `null` inputs, and if you specify a `default_` input value, the schema will use it if the input is `null`. For this reason, the output type may differ from the input type of the schema.

> Note that `nullableAsync` does not accept `undefined` as an input. If you want to accept `undefined` inputs, use <Link href="../optionalAsync/">`optionalAsync`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullishAsync/">`nullishAsync`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallbackAsync/">`fallbackAsync`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nullableAsync` can be used.

##### Nullable username schema

Schema that accepts a unique username or `null`.

> By using a function as the `default_` parameter, the schema will return a unique username from the function call each time the input is `null`.

```ts
import { getUniqueUsername, isUsernameUnique } from '~/api';

const NullableUsernameSchema = v.nullableAsync(
  v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernameUnique, 'The username is not unique.')
  ),
  getUniqueUsername
);
```

##### Unwrap nullable schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nullableAsync`.

```ts
import { isUsernameUnique } from '~/api';

const UsernameSchema = v.unwrap(
  // Assume this schema is from a different file and is reused here
  v.nullableAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `nullableAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### nullishAsync

Creates a nullish schema.

```ts
const Schema = v.nullishAsync<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `nullishAsync` the validation of your schema will pass `undefined` and `null` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined` or `null`. For this reason, the output type may differ from the input type of the schema.

> Note that `nullishAsync` accepts `undefined` or `null` as an input. If you want to accept only `null` inputs, use <Link href="../nullableAsync/">`nullableAsync`</Link>, and if you want to accept only `undefined` inputs, use <Link href="../optionalAsync/">`optionalAsync`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallbackAsync/">`fallbackAsync`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `nullishAsync` can be used.

##### Nullish username schema

Schema that accepts a unique username, `undefined` or `null`.

> By using a function as the `default_` parameter, the schema will return a unique username from the function call each time the input is `undefined` or `null`.

```ts
import { getUniqueUsername, isUsernameUnique } from '~/api';

const NullishUsernameSchema = v.nullishAsync(
  v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernameUnique, 'The username is not unique.')
  ),
  getUniqueUsername
);
```

##### New user schema

Schema to validate new user details.

```ts
import { isEmailUnique, isUsernameUnique } from '~/api';

const NewUserSchema = v.objectAsync({
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailUnique, 'The email is not unique.')
  ),
  username: v.nullishAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

/*
  The input and output types of the schema:
    {
      email: string;
      password: string;
      username?: string | null | undefined;
    }
*/
```

##### Unwrap nullish schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `nullishAsync`.

```ts
import { isUsernameUnique } from '~/api';

const UsernameSchema = v.unwrap(
  // Assume this schema is from a different file and is reused here
  v.nullishAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `nullishAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### objectAsync

Creates an object schema.

```ts
const Schema = v.objectAsync<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `objectAsync` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object, you can use `message` to customize the error message.

> This schema removes unknown entries. The output will only include the entries you specify. To include unknown entries, use <Link href="../looseObjectAsync/">`looseObjectAsync`</Link>. To return an issue for unknown entries, use <Link href="../strictObjectAsync/">`strictObjectAsync`</Link>. To include and validate unknown entries, use <Link href="../objectWithRestAsync/">`objectWithRestAsync`</Link>.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `objectAsync` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### New user schema

Schema to validate an object containing new user details.

```ts
import { isEmailPresent } from '~/api';

const NewUserSchema = v.objectAsync({
  firstName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  lastName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is already in use by another user.')
  ),
  password: v.pipe(v.string(), v.minLength(8)),
  avatar: v.optional(v.pipe(v.string(), v.url())),
});
```

#### Related

The following APIs can be combined with `objectAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={['config', 'getDefault', 'getFallback', 'keyof', 'omit', 'pick']}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### objectWithRestAsync

Creates an object with rest schema.

```ts
const Schema = v.objectWithRestAsync<TEntries, TRest, TMessage>(
  entries,
  rest,
  message
);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `rest` <Property {...properties.rest} />
- `message` <Property {...properties.message} />

##### Explanation

With `objectWithRestAsync` you can validate the data type of the input and whether the content matches `entries` and `rest`. If the input is not an object, you can use `message` to customize the error message.

> The difference to <Link href="../objectAsync/">`objectAsync`</Link> is that this schema includes unknown entries in the output. In addition, this schema filters certain entries from the unknown entries for security reasons.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `objectWithRestAsync` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### Word map schema

Schema to validate an object with word map mutation details.

```ts
import { isUserAllowedToMutate } from '~/api';

// Assume the rest of the keys are always English words
const WordMapSchema = v.objectWithRestAsync(
  {
    $userId: v.pipeAsync(
      v.string(),
      v.regex(/^[a-z0-9]{12}$/i),
      v.checkAsync(
        isUserAllowedToMutate,
        'The user is not allowed to change the word map.'
      )
    ),
    $targetLanguage: v.union([
      v.literal('hindi'),
      v.literal('spanish'),
      v.literal('french'),
    ]),
  },
  v.string()
);
```

#### Related

The following APIs can be combined with `objectWithRestAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={['config', 'getDefault', 'getFallback', 'keyof', 'omit', 'pick']}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### optionalAsync

Creates an optional schema.

```ts
const Schema = v.optionalAsync<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `optionalAsync` the validation of your schema will pass `undefined` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined`. For this reason, the output type may differ from the input type of the schema.

> Note that `optionalAsync` does not accept `null` as an input. If you want to accept `null` inputs, use <Link href="../nullableAsync/">`nullableAsync`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullishAsync/">`nullishAsync`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallbackAsync/">`fallbackAsync`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `optionalAsync` can be used.

##### Optional username schema

Schema that accepts a unique username or `undefined`.

> By using a function as the `default_` parameter, the schema will return a unique username from the function call each time the input is `undefined`.

```ts
import { getUniqueUsername, isUsernameUnique } from '~/api';

const OptionalUsernameSchema = v.optionalAsync(
  v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernameUnique, 'The username is not unique.')
  ),
  getUniqueUsername
);
```

##### New user schema

Schema to validate new user details.

```ts
import { isEmailUnique, isUsernameUnique } from '~/api';

const NewUserSchema = v.objectAsync({
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailUnique, 'The email is not unique.')
  ),
  username: v.optionalAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

/*
  The input and output types of the schema:
    {
      email: string;
      password: string;
      username?: string | undefined;
    }
*/
```

##### Unwrap optional schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `optionalAsync`.

```ts
import { isUsernameUnique } from '~/api';

const UsernameSchema = v.unwrap(
  // Assume this schema is from a different file and is reused here
  v.optionalAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `optionalAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### parseAsync

Parses an unknown input based on a schema.

```ts
const output = v.parseAsync<TSchema>(schema, input, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />
- `config` <Property {...properties.config} />

##### Explanation

`parseAsync` will throw a <Link href="../ValiError/">`ValiError`</Link> if the `input` does not match the `schema`. Therefore you should use a try/catch block to catch errors. If the input matches the schema, it is valid and the `output` of the schema will be returned typed.

> If an asynchronous operation associated with the passed schema throws an error, the promise returned by `parseAsync` is rejected and the error thrown may not be a <Link href="../ValiError/">`ValiError`</Link>.

#### Returns

- `output` <Property {...properties.output} />

#### Examples

The following examples show how `parseAsync` can be used.

```ts
import { isEmailPresent } from '~/api';

try {
  const StoredEmailSchema = v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  );
  const storedEmail = await v.parseAsync(StoredEmailSchema, 'jane@example.com');

  // Handle errors if one occurs
} catch (error) {
  console.error(error);
}
```

#### Related

The following APIs can be combined with `parseAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath', 'isValiError', 'ValiError']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### parserAsync

Returns a function that parses an unknown input based on a schema.

```ts
const parser = v.parserAsync<TSchema, TConfig>(schema, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `config` <Property {...properties.config} />

#### Returns

- `parser` <Property {...properties.parser} />

#### Examples

The following examples show how `parserAsync` can be used.

```ts
import { isEmailPresent } from '~/api';

try {
  const StoredEmailSchema = v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  );
  const storedEmailParser = v.parserAsync(StoredEmailSchema);
  const storedEmail = await storedEmailParser('jane@example.com');

  // Handle errors if one occurs
} catch (error) {
  console.error(error);
}
```

#### Related

The following APIs can be combined with `parserAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath', 'isValiError', 'ValiError']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### partialAsync

Creates a modified copy of an object schema that marks all or only the selected entries as optional.

```ts
const Schema = v.partialAsync<TSchema, TKeys>(schema, keys);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />

##### Explanation

`partialAsync` creates a modified copy of the given object `schema` where all entries or only the selected `keys` are optional. It is similar to TypeScript's [`Partial`](https://www.typescriptlang.org/docs/handbook/utility-types.html#partialtype) utility type.

> Because `partialAsync` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipeAsync/'>`pipeAsync`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipeAsync/'>`pipeAsync`</Link> method after you have modified the schema with `partialAsync`.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `partialAsync` can be used.

##### Update user schema

Schema to update the user details.

```ts
import { isEmailAbsent, isUsernameAbsent } from '~/api';

const UserSchema = v.objectAsync({
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailAbsent, 'The email is already in the database.')
  ),
  username: v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernameAbsent, 'The username is already in the database.')
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

const UpdateUserSchema = v.partialAsync(UserSchema);

/*
  { 
    email?: string;
    username?: string; 
    password?: string;
  }
*/
```

#### Related

The following APIs can be combined with `partialAsync`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'keyof',
    'message',
    'omit',
    'pick',
    'required',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
  ]}
/>

### partialCheckAsync

Creates a partial check validation action.

```ts
const Action = v.partialCheckAsync<TInput, TPaths, TSelection, TMessage>(
  paths,
  requirement,
  message
);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TPaths` <Property {...properties.TPaths} />
- `TSelection` <Property {...properties.TSelection} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `paths` <Property {...properties.paths} />
- `requirement` <Property {...properties.requirement} />
- `message` <Property {...properties.message} />

##### Explanation

With `partialCheckAsync` you can freely validate the selected input and return `true` if it is valid or `false` otherwise. If the input does not match your `requirement`, you can use `message` to customize the error message.

> The difference to <Link href='../checkAsync/'>`checkAsync`</Link> is that `partialCheckAsync` can be executed whenever the selected part of the data is valid, while <Link href='../checkAsync/'>`checkAsync`</Link> is executed only when the entire dataset is typed. This can be an important advantage when working with forms.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `partialCheckAsync` can be used.

##### Message details schema

Schema to validate details associated with a message.

```ts
import { isSenderInTheGroup } from '~/api';

const MessageDetailsSchema = v.pipeAsync(
  v.object({
    sender: v.object({
      name: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
      email: v.pipe(v.string(), v.email()),
    }),
    groupId: v.pipe(v.string(), v.uuid()),
    message: v.pipe(v.string(), v.nonEmpty(), v.maxLength(500)),
  }),
  v.forwardAsync(
    v.partialCheckAsync(
      [['sender', 'email'], ['groupId']],
      (input) =>
        isSenderInTheGroup({
          senderEmail: input.sender.email,
          groupId: input.groupId,
        }),
      'The sender is not in the group.'
    ),
    ['sender', 'email']
  )
);
```

#### Related

The following APIs can be combined with `partialCheckAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'custom',
    'instance',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'object',
    'objectWithRest',
    'record',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'union',
    'variant',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'customAsync',
    'forwardAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'objectAsync',
    'objectWithRestAsync',
    'pipeAsync',
    'recordAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### pipeAsync

Adds a pipeline to a schema, that can validate and transform its input.

```ts
const Schema = v.pipeAsync<TSchema, TItems>(schema, ...items);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TItems` <Property {...properties.TItems} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `items` <Property {...properties.items} />

##### Explanation

`pipeAsync` creates a modified copy of the given `schema`, containing a pipeline for detailed validations and transformations. It passes the input data asynchronously through the `items` in the order they are provided and each item can examine and modify it.

> Since `pipeAsync` returns a schema that can be used as the first argument of another pipeline, it is possible to nest multiple `pipeAsync` calls to extend the validation and transformation further.

`pipeAsync` aborts early and marks the output as untyped if issues were collected before attempting to execute a schema or transformation action as the next item in the pipeline, to prevent unexpected behavior.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `pipeAsync` can be used. Please see the <Link href="/guides/pipelines/">pipeline guide</Link> for more examples and explanations.

##### Stored email schema

Schema to validate a stored email address.

```ts
import { isEmailPresent } from '~/api';

const StoredEmailSchema = v.pipeAsync(
  v.string(),
  v.nonEmpty('Please enter your email.'),
  v.email('The email is badly formatted.'),
  v.maxLength(30, 'Your email is too long.'),
  v.checkAsync(isEmailPresent, 'The email is not in the database.')
);
```

##### New user schema

Schema to validate and transform new user details to a string.

```ts
import { isUsernameUnique } from '~/api';

const NewUserSchema = v.pipeAsync(
  v.objectAsync({
    firstName: v.pipe(v.string(), v.nonEmpty(), v.maxLength(30)),
    lastName: v.pipe(v.string(), v.nonEmpty(), v.maxLength(30)),
    username: v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.maxLength(30),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    ),
  }),
  v.transform(
    ({ firstName, lastName, username }) =>
      `${username} (${firstName} ${lastName})`
  )
);
```

#### Related

The following APIs can be combined with `pipeAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'is',
    'keyof',
    'message',
    'omit',
    'parse',
    'parser',
    'partial',
    'pick',
    'required',
    'safeParse',
    'safeParser',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nanoid',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### rawCheckAsync

Creates a raw check validation action.

```ts
const Action = v.rawCheckAsync<TInput>(action);
```

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Parameters

- `action` <Property {...properties.action} />

##### Explanation

With `rawCheckAsync` you can freely validate the input with a custom `action` and add issues if necessary.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `rawCheckAsync` can be used.

##### Add users schema

Object schema that ensures that only users not already in the group are included.

> This `rawCheckAsync` validation action adds an issue for any invalid username and forwards it via `path` to the appropriate nested field.

```ts
import { isAlreadyInGroup } from '~/api';

const AddUsersSchema = v.pipeAsync(
  v.object({
    groupId: v.pipe(v.string(), v.uuid()),
    usernames: v.array(v.pipe(v.string(), v.nonEmpty())),
  }),
  v.rawCheckAsync(async ({ dataset, addIssue }) => {
    if (dataset.typed) {
      await Promise.all(
        dataset.value.usernames.map(async (username, index) => {
          if (await isAlreadyInGroup(username, dataset.value.groupId)) {
            addIssue({
              received: username,
              message: 'The user is already in the group.',
              path: [
                {
                  type: 'object',
                  origin: 'value',
                  input: dataset.value,
                  key: 'usernames',
                  value: dataset.value.usernames,
                },
                {
                  type: 'array',
                  origin: 'value',
                  input: dataset.value.usernames,
                  key: index,
                  value: username,
                },
              ],
            });
          }
        })
      );
    }
  })
);
```

#### Related

The following APIs can be combined with `rawCheckAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'customAsync',
    'exactOptionalAsync',
    'forwardAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### rawTransformAsync

Creates a raw transformation action.

```ts
const Action = v.rawTransformAsync<TInput, TOutput>(action);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `action` <Property {...properties.action} />

##### Explanation

With `rawTransformAsync` you can freely transform and validate the input with a custom `action` and add issues if necessary.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `rawTransformAsync` can be used.

##### Order schema

Schema that rejects an order that does not meet a requirement when free delivery is expected.

```ts
import { getTotalAmount } from '~/api';
import { FREE_DELIVERY_MIN_AMOUNT } from '~/constants';

const OrderSchema = v.pipeAsync(
  v.object({
    cart: v.array(
      v.object({
        itemId: v.pipe(v.string(), v.uuid()),
        quantity: v.pipe(v.number(), v.integer(), v.minValue(1)),
      })
    ),
    expectsFreeDelivery: v.optional(v.boolean(), false),
  }),
  v.rawTransformAsync(
    async ({ dataset: { value: input }, addIssue, NEVER }) => {
      const total = await getTotalAmount(input.cart);
      if (input.expectsFreeDelivery && total < FREE_DELIVERY_MIN_AMOUNT) {
        addIssue({
          label: 'order',
          expected: `>=${FREE_DELIVERY_MIN_AMOUNT}`,
          received: `${total}`,
          message: `The total amount must be at least $${FREE_DELIVERY_MIN_AMOUNT} for free delivery.`,
          path: [
            {
              type: 'object',
              origin: 'value',
              input,
              key: 'cart',
              value: input.cart,
            },
          ],
        });
        return NEVER;
      }
      return { ...input, total };
    }
  )
);
```

#### Related

The following APIs can be combined with `rawTransformAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'forwardAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### recordAsync

Creates a record schema.

```ts
const Schema = v.recordAsync<TKey, TValue, TMessage>(key, value, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `recordAsync` you can validate the data type of the input and whether the entries match `key` and `value`. If the input is not an object, you can use `message` to customize the error message.

> This schema filters certain entries from the record for security reasons.

> This schema marks an entry as optional if it detects that its key is a literal type. The reason for this is that it is not technically possible to detect missing literal keys without restricting the `key` schema to <Link href="../string/">`string`</Link>, <Link href="../enum/">`enum`</Link> and <Link href="../picklist/">`picklist`</Link>. However, if <Link href="../enum/">`enum`</Link> and <Link href="../picklist/">`picklist`</Link> are used, it is better to use <Link href="../objectAsync/">`objectAsync`</Link> with <Link href="../entriesFromList/">`entriesFromList`</Link> because it already covers the needed functionality. This decision also reduces the bundle size of `recordAsync`, because it only needs to check the entries of the input and not any missing keys.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `recordAsync` can be used.

##### ID to email schema

Schema to validate a record that maps an ID to a public user email.

```ts
import { isEmailPublic } from '~/api';

const IdToEmailSchema = v.recordAsync(
  v.pipe(v.string(), v.uuid()),
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPublic, 'The email address is private.')
  )
);
```

#### Related

The following APIs can be combined with `recordAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### requiredAsync

Creates a modified copy of an object schema that marks all or only the selected entries as required.

```ts
const AllKeysSchema = v.requiredAsync<TSchema, TMessage>(schema, message);
const SelectedKeysSchema = v.requiredAsync<TSchema, TKeys, TMessage>(
  schema,
  keys,
  message
);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TKeys` <Property {...properties.TKeys} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `keys` <Property {...properties.keys} />
- `message` <Property {...properties.message} />

##### Explanation

`requiredAsync` creates a modified copy of the given object `schema` where all or only the selected `keys` are required. It is similar to TypeScript's [`Required`](https://www.typescriptlang.org/docs/handbook/utility-types.html#requiredtype) utility type.

> Because `requiredAsync` changes the data type of the input and output, it is not allowed to pass a schema that has been modified by the <Link href='../pipeAsync/'>`pipeAsync`</Link> method, as this may cause runtime errors. Please use the <Link href='../pipeAsync/'>`pipeAsync`</Link> method after you have modified the schema with `requiredAsync`.

#### Returns

- `AllKeysSchema` <Property {...properties.AllKeysSchema} />
- `SelectedKeysSchema` <Property {...properties.SelectedKeysSchema} />

#### Examples

The following examples show how `requiredAsync` can be used.

##### New task schema

Schema to validate an object containing task details.

```ts
import { isOwnerPresent } from '~/api';

const UpdateTaskSchema = v.objectAsync({
  owner: v.optionalAsync(
    v.pipeAsync(
      v.string(),
      v.email(),
      v.checkAsync(isOwnerPresent, 'The owner is not in the database.')
    )
  ),
  title: v.optional(v.pipe(v.string(), v.nonEmpty(), v.maxLength(255))),
  description: v.optional(v.pipe(v.string(), v.nonEmpty())),
});

const NewTaskSchema = v.requiredAsync(UpdateTaskSchema);

/*
  {
    owner: string;
    title: string;
    description: string;
  }
*/
```

#### Related

The following APIs can be combined with `requiredAsync`.

##### Schemas

<ApiList
  items={[
    'array',
    'exactOptional',
    'intersect',
    'lazy',
    'looseObject',
    'looseTuple',
    'map',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'nullable',
    'nullish',
    'object',
    'objectWithRest',
    'optional',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'tuple',
    'tupleWithRest',
    'undefinedable',
    'union',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'forward',
    'getDefault',
    'getDefaults',
    'getFallback',
    'getFallbacks',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'unwrap',
  ]}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
  ]}
/>

### returnsAsync

Creates a function return transformation action.

```ts
const Action = v.returnsAsync<TInput, TSchema>(schema);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />

##### Explanation

With `returnsAsync` you can force the returned value of a function to match the given `schema`.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `returnsAsync` can be used.

##### Product function schema

Schema of a function that returns a product by its ID.

```ts
import { isValidProductId } from '~/api';

const ProductFunctionSchema = v.pipeAsync(
  v.function(),
  v.argsAsync(
    v.tupleAsync([v.pipeAsync(v.string(), v.checkAsync(isValidProductId))])
  ),
  v.returnsAsync(
    v.pipeAsync(
      v.promise(),
      v.awaitAsync(),
      v.object({
        id: v.string(),
        name: v.string(),
        price: v.number(),
      })
    )
  )
);
```

#### Related

The following APIs can be combined with `returnsAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['pipe']} />

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### safeParseAsync

Parses an unknown input based on a schema.

```ts
const result = v.safeParseAsync<TSchema>(schema, input, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `input` <Property {...properties.input} />
- `config` <Property {...properties.config} />

#### Returns

- `result` <Property {...properties.result} />

#### Example

The following example shows how `safeParseAsync` can be used.

```ts
import { isEmailPresent } from '~/api';

const StoredEmailSchema = v.pipeAsync(
  v.string(),
  v.email(),
  v.checkAsync(isEmailPresent, 'The email is not in the database.')
);
const result = await v.safeParseAsync(StoredEmailSchema, 'jane@example.com');

if (result.success) {
  const storedEmail = result.output;
} else {
  console.error(result.issues);
}
```

#### Related

The following APIs can be combined with `safeParseAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### safeParserAsync

Returns a function that parses an unknown input based on a schema.

```ts
const safeParser = v.safeParserAsync<TSchema, TConfig>(schema, config);
```

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Parameters

- `schema` <Property {...properties.schema} />
- `config` <Property {...properties.config} />

#### Returns

- `safeParser` <Property {...properties.safeParser} />

#### Example

The following example shows how `safeParserAsync` can be used.

```ts
import { isEmailPresent } from '~/api';

const StoredEmailSchema = v.pipeAsync(
  v.string(),
  v.email(),
  v.checkAsync(isEmailPresent, 'The email is not in the database.')
);
const safeStoredEmailParser = v.safeParserAsync(StoredEmailSchema);
const result = await safeStoredEmailParser('jane@example.com');

if (result.success) {
  const storedEmail = result.output;
} else {
  console.error(result.issues);
}
```

#### Related

The following APIs can be combined with `safeParserAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={[
    'assert',
    'config',
    'fallback',
    'flatten',
    'keyof',
    'message',
    'omit',
    'partial',
    'pick',
    'pipe',
    'required',
    'summarize',
    'unwrap',
  ]}
/>

##### Utils

<ApiList items={['getDotPath']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'partialAsync',
    'pipeAsync',
    'recordAsync',
    'requiredAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### setAsync

Creates a set schema.

```ts
const Schema = v.setAsync<TValue, TMessage>(value, message);
```

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `value` <Property {...properties.value} />
- `message` <Property {...properties.message} />

##### Explanation

With `setAsync` you can validate the data type of the input and whether the content matches `value`. If the input is not a set, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `setAsync` can be used.

##### Allowed IPs schema

Schema to validate a set of allowed IP addresses.

```ts
import { isIpAllowed } from '~/api';

const AllowedIPsSchema = v.setAsync(
  v.pipeAsync(
    v.string(),
    v.ip(),
    v.checkAsync(isIpAllowed, 'This IP address is not allowed.')
  )
);
```

#### Related

The following APIs can be combined with `setAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'maxSize',
    'metadata',
    'minSize',
    'notSize',
    'rawCheck',
    'rawTransform',
    'readonly',
    'size',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### strictObjectAsync

Creates a strict object schema.

```ts
const Schema = v.strictObjectAsync<TEntries, TMessage>(entries, message);
```

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `entries` <Property {...properties.entries} />
- `message` <Property {...properties.message} />

##### Explanation

With `strictObjectAsync` you can validate the data type of the input and whether the content matches `entries`. If the input is not an object or does include unknown entries, you can use `message` to customize the error message.

> The difference to <Link href="../objectAsync/">`objectAsync`</Link> is that this schema returns an issue for unknown entries. It intentionally returns only one issue. Otherwise, attackers could send large objects to exhaust device resources. If you want an issue for every unknown key, use the <Link href="../objectWithRestAsync/">`objectWithRestAsync`</Link> schema with <Link href="../never/">`never`</Link> for the `rest` argument.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `strictObjectAsync` can be used. Please see the <Link href="/guides/objects/">object guide</Link> for more examples and explanations.

##### New user schema

Schema to validate a strict object containing only specific new user details.

```ts
import { isEmailPresent } from '~/api';

const NewUserSchema = v.strictObjectAsync({
  firstName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  lastName: v.pipe(v.string(), v.minLength(2), v.maxLength(45)),
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is already in use by another user.')
  ),
  password: v.pipe(v.string(), v.minLength(8)),
  avatar: v.optional(v.pipe(v.string(), v.url())),
});
```

#### Related

The following APIs can be combined with `strictObjectAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList
  items={['config', 'getDefault', 'getFallback', 'keyof', 'omit', 'pick']}
/>

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'forwardAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'requiredAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### strictTupleAsync

Creates a strict tuple schema.

```ts
const Schema = v.strictTupleAsync<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `strictTupleAsync` you can validate the data type of the input and whether the content matches `items`. If the input is not an array or does include unknown items, you can use `message` to customize the error message.

> The difference to <Link href="../tupleAsync/">`tupleAsync`</Link> is that this schema returns an issue for unknown items. It intentionally returns only one issue. Otherwise, attackers could send large arrays to exhaust device resources. If you want an issue for every unknown item, use the <Link href="../tupleWithRestAsync/">`tupleWithRestAsync`</Link> schema with <Link href="../never/">`never`</Link> for the `rest` argument.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `strictTupleAsync` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Number and email tuple

Schema to validate a strict tuple with one number and one stored email address.

```ts
import { isEmailPresent } from '~/api';

const TupleSchema = v.strictTupleAsync([
  v.number(),
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  ),
]);
```

#### Related

The following APIs can be combined with `strictTupleAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### transformAsync

Creates a custom transformation action.

```ts
const Action = v.transformAsync<TInput, TOutput>(operation);
```

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Parameters

- `operation` <Property {...properties.operation} />

##### Explanation

`transformAsync` can be used to freely transform the input. The `operation` parameter is a function that takes the input and returns the transformed output.

#### Returns

- `Action` <Property {...properties.Action} />

#### Examples

The following examples show how `transformAsync` can be used.

##### Blob to string

Schema that transforms a blob to its string value.

```ts
const StringSchema = v.pipeAsync(
  v.blob(),
  v.transformAsync((value) => value.text())
);
```

#### Related

The following APIs can be combined with `transformAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Utils

<ApiList items={['isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'customAsync',
    'exactOptionalAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'pipeAsync',
    'recordAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'undefinedableAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### tupleAsync

Creates a tuple schema.

```ts
const Schema = v.tupleAsync<TItems, TMessage>(items, message);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `message` <Property {...properties.message} />

##### Explanation

With `tupleAsync` you can validate the data type of the input and whether the content matches `items`. If the input is not an array, you can use `message` to customize the error message.

> This schema removes unknown items. The output will only include the items you specify. To include unknown items, use <Link href="../looseTupleAsync/">`looseTupleAsync`</Link>. To return an issue for unknown items, use <Link href="../strictTupleAsync/">`strictTupleAsync`</Link>. To include and validate unknown items, use <Link href="../tupleWithRestAsync/">`tupleWithRestAsync`</Link>.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `tupleAsync` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Number and email tuple

Schema to validate a tuple with one number and one stored email address.

```ts
import { isEmailPresent } from '~/api';

const TupleSchema = v.tupleAsync([
  v.number(),
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  ),
]);
```

#### Related

The following APIs can be combined with `tupleAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### tupleWithRestAsync

Creates a tuple with rest schema.

```ts
const Schema = v.tupleWithRestAsync<TItems, TRest, TMessage>(
  items,
  rest,
  message
);
```

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `items` <Property {...properties.items} />
- `rest` <Property {...properties.rest} />
- `message` <Property {...properties.message} />

##### Explanation

With `tupleWithRestAsync` you can validate the data type of the input and whether the content matches `items` and `rest`. If the input is not an array, you can use `message` to customize the error message.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `tupleWithRestAsync` can be used. Please see the <Link href="/guides/arrays/">arrays guide</Link> for more examples and explanations.

##### Tuple schema with rest

Schema to validate a tuple with generic rest items.

```ts
import { isEmailPresent } from '~/api';

const TupleSchemaWithRest = v.tupleWithRestAsync(
  [
    v.number(),
    v.pipeAsync(
      v.string(),
      v.email(),
      v.checkAsync(isEmailPresent, 'The email is not in the database.')
    ),
  ],
  v.boolean()
);
```

#### Related

The following APIs can be combined with `tupleWithRestAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'check',
    'checkItems',
    'brand',
    'description',
    'empty',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'flavor',
    'includes',
    'length',
    'mapItems',
    'maxLength',
    'metadata',
    'minLength',
    'nonEmpty',
    'notLength',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'someItem',
    'sortItems',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### undefinedableAsync

Creates an undefinedable schema.

```ts
const Schema = v.undefinedableAsync<TWrapped, TDefault>(wrapped, default_);
```

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Parameters

- `wrapped` <Property {...properties.wrapped} />
- `default_` {/* prettier-ignore */}<Property {...properties.default_} />

##### Explanation

With `undefinedableAsync` the validation of your schema will pass `undefined` inputs, and if you specify a `default_` input value, the schema will use it if the input is `undefined`. For this reason, the output type may differ from the input type of the schema.

> `undefinedableAsync` behaves exactly the same as <Link href="../optionalAsync/">`optionalAsync`</Link> at runtime. The only difference is the input and output type when used for object entries. While <Link href="../optionalAsync/">`optionalAsync`</Link> adds a question mark to the key, `undefinedableAsync` does not.

> Note that `undefinedableAsync` does not accept `null` as an input. If you want to accept `null` inputs, use <Link href="../nullableAsync/">`nullableAsync`</Link>, and if you want to accept `null` and `undefined` inputs, use <Link href="../nullishAsync/">`nullishAsync`</Link> instead. Also, if you want to set a default output value for any invalid input, you should use <Link href="../fallbackAsync/">`fallbackAsync`</Link> instead.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `undefinedableAsync` can be used.

##### Undefinedable username schema

Schema that accepts a unique username or `undefined`.

> By using a function as the `default_` parameter, the schema will return a unique username from the function call each time the input is `undefined`.

```ts
import { getUniqueUsername, isUsernameUnique } from '~/api';

const UndefinedableUsernameSchema = v.undefinedableAsync(
  v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernameUnique, 'The username is not unique.')
  ),
  getUniqueUsername
);
```

##### New user schema

Schema to validate new user details.

```ts
import { isEmailUnique, isUsernameUnique } from '~/api';

const NewUserSchema = v.objectAsync({
  email: v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailUnique, 'The email is not unique.')
  ),
  username: v.undefinedableAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  ),
  password: v.pipe(v.string(), v.minLength(8)),
});

/*
  The input and output types of the schema:
    {
      email: string;
      password: string;
      username: string | undefined;
    }
*/
```

##### Unwrap undefinedable schema

Use <Link href="../unwrap/">`unwrap`</Link> to undo the effect of `undefinedableAsync`.

```ts
import { isUsernameUnique } from '~/api';

const UsernameSchema = v.unwrap(
  // Assume this schema is from a different file and is reused here
  v.undefinedableAsync(
    v.pipeAsync(
      v.string(),
      v.nonEmpty(),
      v.checkAsync(isUsernameUnique, 'The username is not unique.')
    )
  )
);
```

#### Related

The following APIs can be combined with `undefinedableAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonUndefinedable',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback', 'unwrap']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'flavor',
    'metadata',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'unionAsync',
    'variantAsync',
  ]}
/>

### unionAsync

Creates an union schema.

> I recommend that you read the <Link href="/guides/unions/">unions guide</Link> before using this schema function.

```ts
const Schema = v.unionAsync<TOptions, TMessage>(options, message);
```

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `unionAsync` you can validate if the input matches one of the given `options`. If the input does not match a schema and cannot be clearly assigned to one of the options, you can use `message` to customize the error message.

If a bad input can be uniquely assigned to one of the schemas based on the data type, the result of that schema is returned. Otherwise, a general issue is returned that contains the issues of each schema as subissues. This is a special case within the library, as the issues of `unionAsync` can contradict each other.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `unionAsync` can be used.

##### User schema

Schema to validate a user's email or username.

```ts
import { isEmailPresent, isUsernamePresent } from '~/api';

const UserSchema = v.unionAsync([
  v.pipeAsync(
    v.string(),
    v.email(),
    v.checkAsync(isEmailPresent, 'The email is not in the database.')
  ),
  v.pipeAsync(
    v.string(),
    v.nonEmpty(),
    v.checkAsync(isUsernamePresent, 'The username is not in the database.')
  ),
]);
```

#### Related

The following APIs can be combined with `unionAsync`.

##### Schemas

<ApiList
  items={[
    'any',
    'array',
    'bigint',
    'blob',
    'boolean',
    'custom',
    'date',
    'enum',
    'exactOptional',
    'file',
    'function',
    'instance',
    'intersect',
    'lazy',
    'literal',
    'looseObject',
    'looseTuple',
    'map',
    'nan',
    'never',
    'nonNullable',
    'nonNullish',
    'nonOptional',
    'null',
    'nullable',
    'nullish',
    'number',
    'object',
    'objectWithRest',
    'optional',
    'picklist',
    'promise',
    'record',
    'set',
    'strictObject',
    'strictTuple',
    'string',
    'symbol',
    'tuple',
    'tupleWithRest',
    'undefined',
    'undefinedable',
    'union',
    'unknown',
    'variant',
    'void',
  ]}
/>

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'args',
    'base64',
    'bic',
    'brand',
    'bytes',
    'check',
    'checkItems',
    'creditCard',
    'cuid2',
    'decimal',
    'description',
    'digits',
    'email',
    'emoji',
    'empty',
    'endsWith',
    'entries',
    'everyItem',
    'excludes',
    'filterItems',
    'findItem',
    'finite',
    'flavor',
    'graphemes',
    'gtValue',
    'hash',
    'hexadecimal',
    'hexColor',
    'imei',
    'includes',
    'integer',
    'ip',
    'ipv4',
    'ipv6',
    'isoDate',
    'isoDateTime',
    'isoTime',
    'isoTimeSecond',
    'isoTimestamp',
    'isoWeek',
    'length',
    'ltValue',
    'mac',
    'mac48',
    'mac64',
    'mapItems',
    'maxBytes',
    'maxEntries',
    'maxGraphemes',
    'maxLength',
    'maxSize',
    'maxValue',
    'maxWords',
    'metadata',
    'mimeType',
    'minBytes',
    'minEntries',
    'minGraphemes',
    'minLength',
    'minSize',
    'minValue',
    'minWords',
    'multipleOf',
    'nonEmpty',
    'notBytes',
    'notEntries',
    'notGraphemes',
    'notLength',
    'notSize',
    'notValue',
    'notValues',
    'notWords',
    'octal',
    'parseJson',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'reduceItems',
    'regex',
    'returns',
    'rfcEmail',
    'safeInteger',
    'size',
    'slug',
    'someItem',
    'sortItem',
    'startsWith',
    'stringifyJson',
    'title',
    'toLowerCase',
    'toMaxValue',
    'toMinValue',
    'toUpperCase',
    'transform',
    'trim',
    'trimEnd',
    'trimStart',
    'ulid',
    'url',
    'uuid',
    'value',
    'values',
    'words',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'arrayAsync',
    'awaitAsync',
    'checkAsync',
    'customAsync',
    'exactOptionalAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'intersectAsync',
    'lazyAsync',
    'looseObjectAsync',
    'looseTupleAsync',
    'mapAsync',
    'nonNullableAsync',
    'nonNullishAsync',
    'nonOptionalAsync',
    'nullableAsync',
    'nullishAsync',
    'objectAsync',
    'objectWithRestAsync',
    'optionalAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'recordAsync',
    'safeParseAsync',
    'safeParserAsync',
    'setAsync',
    'strictObjectAsync',
    'strictTupleAsync',
    'transformAsync',
    'tupleAsync',
    'tupleWithRestAsync',
    'variantAsync',
  ]}
/>

### variantAsync

Creates a variant schema.

```ts
const Schema = v.variantAsync<TKey, TOptions, TMessage>(key, options, message);
```

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Parameters

- `key` <Property {...properties.key} />
- `options` <Property {...properties.options} />
- `message` <Property {...properties.message} />

##### Explanation

With `variantAsync` you can validate if the input matches one of the given object `options`. The object schema to be used for the validation is determined by the discriminator `key`. If the input does not match a schema and cannot be clearly assigned to one of the options, you can use `message` to customize the error message.

> It is allowed to specify the exact same or a similar discriminator multiple times. However, in such cases `variantAsync` will only return the output of the first untyped or typed variant option result. Typed results take precedence over untyped ones.

> For deeply nested `variant` schemas with several different discriminator keys, `variant` will return an issue for the first most likely object schemas on invalid input. The order of the discriminator keys and the presence of a discriminator in the input are taken into account.

#### Returns

- `Schema` <Property {...properties.Schema} />

#### Examples

The following examples show how `variantAsync` can be used.

##### Message schema

Schema to validate a message object.

```ts
import { isValidGroupReceiver, isValidUserReceiver } from '~/api';

const MessageSchema = v.objectAsync({
  message: v.pipe(v.string(), v.nonEmpty()),
  receiver: v.variantAsync('type', [
    v.objectAsync({
      type: v.literal('group'),
      groupId: v.pipeAsync(
        v.string(),
        v.uuid(),
        v.checkAsync(isValidGroupReceiver, 'The group cannot receive messages.')
      ),
    }),
    v.objectAsync({
      type: v.literal('user'),
      email: v.pipeAsync(
        v.string(),
        v.email(),
        v.checkAsync(isValidUserReceiver, 'The user cannot receive messages.')
      ),
    }),
  ]),
});
```

##### User schema

Schema to validate unique user details.

```ts
import { isRegisteredEmail, isRegisteredUsername, isValidUserId } from '~/api';

const UserSchema = v.variantAsync('type', [
  // Assume this schema is from a different file and reused here.
  v.variantAsync('type', [
    v.objectAsync({
      type: v.literal('email'),
      email: v.pipeAsync(
        v.string(),
        v.email(),
        v.checkAsync(isRegisteredEmail, 'The email is not registered.')
      ),
    }),
    v.objectAsync({
      type: v.literal('username'),
      username: v.pipeAsync(
        v.string(),
        v.nonEmpty(),
        v.checkAsync(isRegisteredUsername, 'The username is not registered.')
      ),
    }),
  ]),
  v.objectAsync({
    type: v.literal('userId'),
    userId: v.pipeAsync(
      v.string(),
      v.uuid(),
      v.checkAsync(isValidUserId, 'The user id is not valid.')
    ),
  }),
]);
```

#### Related

The following APIs can be combined with `variantAsync`.

##### Schemas

<ApiList items={['looseObject', 'object', 'objectWithRest', 'strictObject']} />

##### Methods

<ApiList items={['config', 'getDefault', 'getFallback']} />

##### Actions

<ApiList
  items={[
    'brand',
    'check',
    'description',
    'entries',
    'flavor',
    'maxEntries',
    'metadata',
    'minEntries',
    'notEntries',
    'partialCheck',
    'rawCheck',
    'rawTransform',
    'readonly',
    'title',
    'transform',
  ]}
/>

##### Utils

<ApiList items={['entriesFromList', 'isOfKind', 'isOfType']} />

##### Async

<ApiList
  items={[
    'checkAsync',
    'fallbackAsync',
    'getDefaultsAsync',
    'getFallbacksAsync',
    'looseObjectAsync',
    'objectAsync',
    'objectWithRestAsync',
    'parseAsync',
    'parserAsync',
    'partialCheckAsync',
    'pipeAsync',
    'rawCheckAsync',
    'rawTransformAsync',
    'safeParseAsync',
    'safeParserAsync',
    'strictObjectAsync',
    'transformAsync',
  ]}
/>

## Types (API)

### AnySchema

Any schema interface.

#### Definition

- `AnySchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />

### ArgsAction

Args action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `ArgsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `schema` <Property {...properties.schema} />

### ArgsActionAsync

Args action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `ArgsActionAsync` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `schema` <Property {...properties.schema} />

### ArrayInput

Array input type.

#### Definition

- `ArrayInput` <Property {...properties.ArrayInput} />

### ArrayIssue

Array issue interface.

#### Definition

- `ArrayIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### ArrayPathItem

Array path item interface.

#### Definition

- `ArrayPathItem`
  - `type` <Property {...properties.type} />
  - `origin` <Property {...properties.origin} />
  - `input` <Property {...properties.input} />
  - `key` <Property type="number" />
  - `value` <Property type="unknown" />

The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

### ArrayRequirement

Array requirement type.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ArrayRequirement` <Property {...properties.ArrayRequirement} />

### ArrayRequirementAsync

Array requirement async type.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ArrayRequirementAsync` <Property {...properties.ArrayRequirementAsync} />

### ArraySchema

Array schema interface.

#### Generics

- `TItem` <Property {...properties.TItem} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ArraySchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `item` <Property {...properties.item} />
  - `message` <Property {...properties.message} />

### ArraySchemaAsync

Array schema async interface.

#### Generics

- `TItem` <Property {...properties.TItem} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ArraySchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `item` <Property {...properties.item} />
  - `message` <Property {...properties.message} />

### AwaitActionAsync

Await action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `AwaitActionAsync` <Property {...properties.BaseTransformationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### Base64Action

Base64 action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Base64Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Base64Issue

Base64 issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Base64Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### BaseIssue

Schema issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `BaseIssue` <Property {...properties.Config} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `input` <Property {...properties.input} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `message` <Property {...properties.message} />
  - `requirement` <Property {...properties.requirement} />
  - `path` <Property {...properties.path} />
  - `issues` <Property {...properties.issues} />

### BaseMetadata

Base metadata interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `BaseMetadata`
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `~types` <Property {...properties['~types']} />

### BaseSchema

Base schema interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseSchema`
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `async` <Property {...properties.async} />
  - `~standard` <Property {...properties['~standard']} />
  - `~run` <Property {...properties['~run']} />
  - `~types` <Property {...properties['~types']} />

### BaseSchemaAsync

Base schema async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `reference` <Property {...properties.reference} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />

### BaseTransformation

Base transformation interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseTransformation`
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />
  - `~types` <Property {...properties['~types']} />

### BaseTransformationAsync

Base transformation async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseTransformationAsync` <Property {...properties.BaseTransformationAsync} />
  - `reference` <Property {...properties.reference} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />

### BaseValidation

Base action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseValidation`
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />
  - `~types` <Property {...properties['~types']} />

### BaseValidationAsync

Base validation async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `BaseValidationAsync` <Property {...properties.BaseValidationAsync} />
  - `reference` <Property {...properties.reference} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />

### BicAction

BIC action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `BicAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### BicIssue

Bic issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `BicIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### BigintIssue

Bigint issue interface.

#### Definition

- `BigintIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### BigintSchema

Bigint schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `BigintSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### BlobIssue

Blob issue interface.

#### Definition

- `BlobIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### BlobSchema

Blob schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `BlobSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### BooleanIssue

Boolean issue interface.

#### Definition

- `BooleanIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### BooleanSchema

Boolean schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `BooleanSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### Brand

Brand interface.

#### Generics

- `TName` <Property {...properties.TName} />

#### Definition

- `Brand` <Property {...properties.Brand} />

### BrandAction

Brand action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TName` <Property {...properties.TName} />

#### Definition

- `BrandAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `name` <Property {...properties.name} />

### BrandName

Brand name type.

#### Definition

- `BrandName` <Property {...properties.BrandName} />

### BytesAction

Bytes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `BytesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### BytesIssue

Bytes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `BytesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### CheckAction

Check action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CheckAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### CheckActionAsync

Check action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CheckActionAsync` <Property {...properties.BaseValidationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### CheckIssue

Check issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `CheckIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### CheckItemsAction

Check items action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CheckItemsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### CheckItemsActionAsync

Check items action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CheckItemsActionAsync` <Property {...properties.BaseValidationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### CheckItemsIssue

Check items issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `CheckItemsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### Class

Class type.

#### Definition

- `Class` <Property {...properties.Class} />

### Config

Config interface.

#### Generics

- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `Config`
  - `lang` <Property {...properties.lang} />
  - `message` <Property {...properties.message} />
  - `abortEarly` <Property {...properties.abortEarly} />
  - `abortPipeEarly` <Property {...properties.abortPipeEarly} />

### ContentInput

Content input type.

#### Definition

- `ContentInput` <Property {...properties.ContentInput} />

### ContentRequirement

Content requirement type.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ContentRequirement` <Property {...properties.ContentRequirement} />

### CreditCardAction

Credit card action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CreditCardAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### CreditCardIssue

Credit card issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `CreditCardIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Cuid2Action

Cuid2 action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Cuid2Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Cuid2Issue

Cuid2 issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Cuid2Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### CustomIssue

Custom issue interface.

#### Definition

- `CustomIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### CustomSchema

Custom schema interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CustomSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `check` <Property {...properties.check} />
  - `message` <Property {...properties.message} />

### CustomSchemaAsync

Custom schema async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `CustomSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `check` <Property {...properties.check} />
  - `message` <Property {...properties.message} />

### DateIssue

Date issue interface.

#### Definition

- `DateIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### DateSchema

Date schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `DateSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### DecimalAction

Decimal action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `DecimalAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### DecimalIssue

Decimal issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `DecimalIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### DeepPickN

Deeply picks N specific keys.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/actions/partialCheck/types.ts).

### Default

Default type.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TInput` <Property {...properties.TInput} />

#### Definition

- `Default` <Property {...properties.Default} />

### DefaultAsync

Default async type.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TInput` <Property {...properties.TInput} />

#### Definition

- `DefaultAsync` <Property {...properties.DefaultAsync} />

### DefaultValue

Default value type.

#### Generics

- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `DefaultValue` <Property {...properties.DefaultValue} />

### DescriptionAction

Description action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TDescription` <Property {...properties.TDescription} />

#### Definition

- `DescriptionAction` <Property {...properties.BaseMetadata} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `description` <Property {...properties.description} />

### DigitsAction

Digits action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `DigitsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### DigitsIssue

Digits issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `DigitsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### EmailAction

Email action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EmailAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### EmailIssue

Email issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `EmailIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### EmojiAction

Emoji action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EmojiAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### EmojiIssue

Emoji issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `EmojiIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### EmptyAction

Empty action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EmptyAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### EmptyIssue

Empty issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `EmptyIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />

### EndsWithAction

Ends with action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EndsWithAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### EndsWithIssue

Ends with issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `EndsWithIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### EntriesAction

Entries action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EntriesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### EntriesInput

Entries input type.

#### Definition

- `EntriesInput` <Property {...properties.EntriesInput} />

### EntriesIssue

Entries issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `EntriesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Enum

Enum interface.

#### Definition

- `Enum` <Property {...properties.Enum} />

### EnumIssue

Enum issue interface.

#### Definition

- `EnumIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### EnumSchema

Enum schema interface.

#### Generics

- `TEnum` <Property {...properties.TEnum} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EnumSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `enum` <Property {...properties.enum} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### ErrorMessage

Error message type.

#### Generics

- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `ErrorMessage` <Property {...properties.ErrorMessage} />

### EveryItemAction

Every action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EveryItemAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### EveryItemIssue

Every item issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `EveryItemIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### ExactOptionalSchema

Exact optional schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `ExactOptionalSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### ExactOptionalSchemaAsync

Exact optional schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `ExactOptionalSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### ExcludesAction

Excludes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ExcludesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `referece` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### ExcludesIssue

Excludes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `ExcludesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### FailureDataset

Failure dataset interface.

#### Generics

- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `UntypedDataset`
  - `typed` <Property {...properties.typed} />
  - `value` <Property {...properties.value} />
  - `issues` <Property {...properties.issues} />

### Fallback

Fallback type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `Fallback` <Property {...properties.Fallback} />

### FallbackAsync

Fallback async type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `FallbackAsync` <Property {...properties.FallbackAsync} />

### FileIssue

File issue interface.

#### Definition

- `FileIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### FileSchema

File schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `FileSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### FilterItemsAction

Filter items action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `FilterItemsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### FindItemAction

Find item action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `FindItemAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### FiniteAction

Finite action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `FiniteAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### FiniteIssue

Finite issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `FiniteIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### FirstTupleItem

Extracts first tuple item.

#### Generics

- `TTuple` <Property {...properties.TTuple} />

#### Definition

- `FirstTupleItem` <Property {...properties.FirstTupleItem} />

### FlatErrors

Flat errors type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/flatten/flatten.ts).

### Flavor

Flavor interface.

#### Generics

- `TName` <Property {...properties.TName} />

#### Definition

- `Flavor` <Property {...properties.Flavor} />

### FlavorAction

Flavor action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TName` <Property {...properties.TName} />

#### Definition

- `FlavorAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `name` <Property {...properties.name} />

### FlavorName

Flavor name type.

#### Definition

- `FlavorName` <Property {...properties.FlavorName} />

### FunctionIssue

Function issue interface.

#### Definition

- `FunctionIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### FunctionSchema

Function schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `FunctionSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### GenericIssue

Generic issue type.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `GenericIssue` <Property {...properties.BaseIssue} />

### GenericMetadata

Generic metadata type.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `GenericMetadata` <Property {...properties.BaseMetadata} />

### GenericPipeAction

Generic pipe action type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericPipeAction` <Property {...properties.PipeAction} />

### GenericPipeActionAsync

Generic pipe action async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericPipeActionAsync` <Property {...properties.PipeActionAsync} />

### GenericPipeItem

Generic pipe item type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericPipeItem` <Property {...properties.PipeItem} />

### GenericPipeItemAsync

Generic pipe item async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericPipeItemAsync` <Property {...properties.PipeItemAsync} />

### GenericSchema

Generic schema type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericSchema` <Property {...properties.BaseSchema} />

### GenericSchemaAsync

Generic schema async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericSchemaAsync` <Property {...properties.BaseSchemaAsync} />

### GenericTransformation

Generic transformation type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericTransformation` <Property {...properties.BaseTransformation} />

### GenericTransformationAsync

Generic transformation async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericTransformationAsync` <Property {...properties.BaseTransformationAsync} />

### GenericValidation

Generic validation type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericValidation` <Property {...properties.BaseValidation} />

### GenericValidationAsync

Generic validation async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `GenericValidationAsync` <Property {...properties.BaseValidationAsync} />

### GlobalConfig

The global config type.

#### Definition

- `GlobalConfig` <Property {...properties.GlobalConfig} />

### GraphemesAction

Graphemes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `GraphemesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### GraphemesIssue

Graphemes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `GraphemesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### GtValueAction

Greater than value action type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `GtValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### GtValueIssue

Greater than value issue type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `GtValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### HashAction

Hash action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `HashAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### HashIssue

Hash issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `HashIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### HashType

Hash type type.

#### Definition

- `HashType` <Property {...properties.HashType} />

### HexadecimalAction

Hexadecimal action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `HexadecimalAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### HexadecimalIssue

Hexadecimal issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `HexadecimalIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### HexColorAction

Hex color action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `HexColorAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### HexColorIssue

HexColor issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `HexColorIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### ImeiAction

Imei action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ImeiAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### ImeiIssue

IMEI issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ImeiIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IncludesAction

Includes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IncludesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IncludesIssue

Includes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `IncludesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### InferDefault

Infer default type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/getDefault/getDefault.ts).

### InferDefaults

Infer defaults type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/getDefaults/types.ts).

### InferFallback

Infer fallback type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/getFallback/getFallback.ts).

### InferFallbacks

Infer fallbacks type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/getFallbacks/types.ts).

### InferInput

Infer input type.

#### Generics

- `TItem` <Property {...properties.TItem} />

#### Definition

- `InferInput` <Property {...properties.InferInput} />

#### Example

```ts
// Create object schema
const ObjectSchema = v.object({
  key: v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
});

// Infer object input type
type ObjectInput = v.InferInput<typeof ObjectSchema>; // { key: string }
```

### InferIntersectInput

Infer intersect input type.

```ts
// Create object schemas
const ObjectSchemas = [
  v.object({
    key1: v.pipe(
      v.string(),
      v.transform((input) => input.length)
    ),
  }),
  v.object({
    key2: v.pipe(
      v.string(),
      v.transform((input) => input.length)
    ),
  }),
];

// Infer object intersect input type
type ObjectInput = v.InferIntersectInput<typeof ObjectSchemas>; // { key1: string } & { key2: string }
```

### InferIntersectOutput

Infer intersect output type.

```ts
// Create object schemas
const ObjectSchemas = [
  v.object({
    key1: v.pipe(
      v.string(),
      v.transform((input) => input.length)
    ),
  }),
  v.object({
    key2: v.pipe(
      v.string(),
      v.transform((input) => input.length)
    ),
  }),
];

// Infer object intersect output type
type ObjectOutput = v.InferIntersectOutput<typeof ObjectSchemas>; // { key1: number } & { key2: number }
```

### InferIssue

Infer issue type.

#### Generics

- `TItem` <Property {...properties.TItem} />

#### Definition

- `InferIssue` <Property {...properties.InferIssue} />

### InferMapInput

Infer map input type.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />

#### Definition

- `InferMapInput` <Property {...properties.InferMapInput} />

### InferMapOutput

Infer map output type.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />

#### Definition

- `InferMapOutput` <Property {...properties.InferMapOutput} />

### InferMetadata

Infer fallbacks type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/getMetadata/getMetadata.ts).

### InferNonNullableInput

Infer non nullable input type.

```ts
// Create nullable string schema
const NullableStringSchema = v.nullable(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non nullable string input type
type NonNullableStringInput = v.InferNonNullableInput<
  typeof NullableStringSchema
>; // string
```

### InferNonNullableIssue

Infer non nullable issue type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/schemas/nonNullable/types.ts).

### InferNonNullableOutput

Infer non nullable output type.

```ts
// Create nullable string schema
const NullableStringSchema = v.nullable(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non nullable string output type
type NonNullableStringOutput = v.InferNonNullableOutput<
  typeof NullableStringSchema
>; // number
```

### InferNonNullishInput

Infer non nullable input type.

```ts
// Create nullish string schema
const NullishStringSchema = v.nullish(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non nullish string input type
type NonNullishStringInput = v.InferNonNullishInput<typeof NullishStringSchema>; // string
```

### InferNonNullishIssue

Infer non nullish issue type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/schemas/nonNullish/types.ts).

### InferNonNullishOutput

Infer non nullable output type.

```ts
// Create nullish string schema
const NullishStringSchema = v.nullish(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non nullish string output type
type NonNullishStringOutput = v.InferNonNullishOutput<
  typeof NullishStringSchema
>; // number
```

### InferNonOptionalInput

Infer non optional input type.

```ts
// Create optional string schema
const OptionalStringSchema = v.optional(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non optional string input type
type NonOptionalStringInput = v.InferNonOptionalInput<
  typeof OptionalStringSchema
>; // string
```

### InferNonOptionalIssue

Infer non optional issue type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/schemas/nonOptional/types.ts).

### InferNonOptionalOutput

Infer non optional output type.

```ts
// Create optional string schema
const OptionalStringSchema = v.optional(
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  )
);

// Infer non optional string output type
type NonOptionalStringOutput = v.InferNonOptionalOutput<
  typeof OptionalStringSchema
>; // number
```

### InferNullableOutput

Infer nullable output type.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `InferNullableOutput` <Property {...properties.InferNullableOutput} />

### InferNullishOutput

Infer nullish output type.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `InferNullishOutput` <Property {...properties.InferNullishOutput} />

### InferObjectInput

Infer object input type.

```ts
// Create object entries
const entries = {
  key: v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
};

// Infer entries input type
type EntriesInput = v.InferObjectInput<typeof entries>; // { key: string }
```

### InferObjectIssue

Infer object issue type.

#### Generics

- `TEntries` <Property {...properties.TEntries} />

#### Definition

- `InferObjectIssue` <Property {...properties.InferObjectIssue} />

### InferObjectOutput

Infer object output type.

```ts
// Create object entries
const entries = {
  key: v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
};

// Infer entries output type
type EntriesOutput = v.InferObjectOutput<typeof entries>; // { key: number }
```

### InferOptionalOutput

Infer optional output type.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `InferOptionalOutput` <Property {...properties.InferOptionalOutput} />

### InferOutput

Infer output type.

#### Generics

- `TItem` <Property {...properties.TItem} />

#### Definition

- `InferIssue` <Property {...properties.InferIssue} />

#### Example

```ts
// Create object schema
const ObjectSchema = v.object({
  key: v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
});

// Infer object output type
type ObjectOutput = v.InferOutput<typeof ObjectSchema>; // { key: number }
```

### InferRecordInput

Infer record input type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/schemas/record/types.ts).

### InferRecordOutput

Infer record output type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/schemas/record/types.ts).

### InferSetInput

Infer set input type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `InferSetInput` <Property {...properties.InferSetInput} />

### InferSetOutput

Infer set output type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `InferSetOutput` <Property {...properties.InferSetOutput} />

### InferTupleInput

Infer tuple output type.

```ts
// Create tuple items
const items = [
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
];

// Infer items input type
type ItemsInput = v.InferTupleInput<typeof items>; // [string]
```

### InferTupleIssue

Infer tuple issue type.

#### Generics

- `TItems` <Property {...properties.TItems} />

#### Definition

- `InferTupleIssue` <Property {...properties.InferTupleIssue} />

### InferTupleOutput

Infer tuple issue type.

```ts
const items = [
  v.pipe(
    v.string(),
    v.transform((input) => input.length)
  ),
];

// Infer items output type
type ItemsOutput = v.InferTupleOutput<typeof items>; // [number]
```

### InferVariantIssue

Infer variant issue type.

#### Generics

- `TOptions` <Property {...properties.TOptions} />

#### Definition

- `InferVariantIssue` <Property {...properties.InferVariantIssue} />

### InstanceIssue

Instance issue interface.

#### Definition

- `InstanceIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### InstanceSchema

Instance schema interface.

#### Generics

- `TClass` <Property {...properties.TClass} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `InstanceSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `class` <Property {...properties.class} />
  - `message` <Property {...properties.message} />

### IntegerAction

Integer action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IntegerAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IntegerIssue

Integer issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IntegerIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IntersectIssue

Intersect issue interface.

#### Definition

- `IntersectIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### IntersectOptions

Intersect options type.

#### Definition

- `IntersectOptions` <Property {...properties.IntersectOptions} />

### IntersectOptionsAsync

Intersect options async type.

#### Definition

- `IntersectOptionsAsync` <Property {...properties.IntersectOptionsAsync} />

### IntersectSchema

Intersect schema interface.

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IntersectSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### IntersectSchemaAsync

Intersect schema async interface.

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IntersectSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### IpAction

IP action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IpAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IpIssue

IP issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IpIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Ipv4Action

IPv4 action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Ipv4Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Ipv4Issue

IPv4 issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Ipv4Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Ipv6Action

IPv6 action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Ipv6Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Ipv6Issue

IPv6 issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Ipv6Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoDateAction

ISO date action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoDateAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoDateIssue

ISO date issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoDateIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoDateTimeAction

ISO date time action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoDateTimeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoDateTimeIssue

ISO date time issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoDateTimeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoTimeAction

ISO time action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoTimeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoTimeIssue

ISO time issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoTimeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoTimeSecondAction

ISO time second action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoTimeSecondAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoTimeSecondIssue

ISO time second issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoTimeSecondIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoTimestampAction

ISO timestamp action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoTimestampAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoTimestampIssue

ISO timestamp issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoTimestampIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IsoWeekAction

ISO week action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `IsoWeekAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### IsoWeekIssue

ISO week issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `IsoWeekIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### IssueDotPath

Issue dot path type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/types/issue.ts).

### IssuePathItem

Path item type.

#### Definition

- `IssuePathItem` <Property {...properties.IssuePathItem} />

### LazySchema

Lazy schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />

#### Definition

- `LazySchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `getter` <Property {...properties.getter} />

### LazySchemaAsync

Lazy schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />

#### Definition

- `LazySchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `getter` <Property {...properties.getter} />

### LengthAction

Length action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LengthAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### LengthInput

Length input type.

#### Definition

- `LengthInput` <Property {...properties.LengthInput} />

### LengthIssue

Length issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `LengthIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Literal

Literal type.

#### Definition

- `Literal` <Property {...properties.Literal} />

### LiteralIssue

Literal issue interface.

#### Definition

- `LiteralIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### LooseObjectIssue

Loose object issue interface.

#### Definition

- `LooseObjectIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### LooseObjectSchema

Loose object schema interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LooseObjectSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### LooseObjectSchemaAsync

Loose object schema async interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LooseObjectSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### LooseTupleIssue

Loose tuple issue interface.

#### Definition

- `LooseTupleIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### LooseTupleSchema

Loose tuple schema interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LooseTupleSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### LooseTupleSchemaAsync

Loose tuple schema async interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LooseTupleSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### LiteralSchema

Literal schema interface.

#### Generics

- `TLiteral` <Property {...properties.TLiteral} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LiteralSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `literal` <Property {...properties.literal} />
  - `message` <Property {...properties.message} />

### LtValueAction

Less than value action type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `LtValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### LtValueIssue

Less than value issue type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `LtValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### Mac48Action

48-bit MAC action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Mac48Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Mac48Issue

48-bit MAC issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Mac48Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### Mac64Action

64-bit MAC action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `Mac64Action` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### Mac64Issue

64-bit MAC issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `Mac64Issue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MacAction

MAC action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MacAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MacIssue

MAC issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `MacIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MapIssue

Map issue interface.

#### Definition

- `MapIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### MapItemsAction

Map items action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `MapItemsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### MapPathItem

Map path item interface.

#### Definition

- `MapPathItem`
  - `type` <Property {...properties.type} />
  - `origin` <Property {...properties.origin} />
  - `input` <Property {...properties.input} />
  - `key` <Property type='unknown' />
  - `value` <Property type='unknown' />

The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

### MapSchema

Map schema interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MapSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### MapSchemaAsync

Map schema async interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MapSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### MaxBytesAction

Max bytes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxBytesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxBytesIssue

Max bytes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxBytesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaxEntriesAction

Max entries action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxEntriesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxEntriesIssue

Max entries issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxEntriesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaxGraphemesAction

Max graphemes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxGraphemesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxGraphemesIssue

Max graphemes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxGraphemesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaxLengthAction

Max length action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxLengthAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxLengthIssue

Max length issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxLengthIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaxSizeAction

Max size action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxSizeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxSizeIssue

Max size issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxSizeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaxValueAction

Max value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxValueIssue

Max value issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### MaxWordsAction

Max words action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MaxWordsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `locales` <Property {...properties.locales} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MaxWordsIssue

Max words issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MaxWordsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MaybePromise

Maybe promise type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `MaybePromise` <Property {...properties.MaybePromise} />

### MaybeReadonly

Maybe readonly type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `MaybeReadonly` <Property {...properties.MaybeReadonly} />

### MetadataAction

Metadata action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMetadata` <Property {...properties.TMetadata} />

#### Definition

- `MetadataAction` <Property {...properties.BaseMetadata} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `metadata_` <Property {...properties['metadata_']} />

### MimeTypeAction

MIME type action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MimeTypeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MimeTypeIssue

Mime type issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MimeTypeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinBytesAction

Min bytes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinBytesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinBytesIssue

Min bytes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinBytesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinEntriesAction

Min entries action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinEntriesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinEntriesIssue

Min entries issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinEntriesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinGraphemesAction

Min graphemes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinGraphemesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinGraphemesIssue

Min graphemes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinGraphemesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinLengthAction

Min length action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinLengthAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinLengthIssue

Min length issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinLengthIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinSizeAction

Min size action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinSizeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `referece` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinSizeIssue

Min size issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinSizeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MinValueAction

Min value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinValueIssue

Min value issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### MinWordsAction

Min words action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MinWordsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `locales` <Property {...properties.locales} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MinWordsIssue

Min words issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MinWordsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### MultipleOfAction

Multiple of action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `MultipleOfAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### MultipleOfIssue

Multiple of issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `MultipleOfIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NanIssue

NaN issue interface.

#### Definition

- `NanIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NanSchema

NaN schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NanSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### NeverIssue

Never issue interface.

#### Definition

- `NeverIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NeverSchema

Never schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NeverSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### NonEmptyAction

Non empty action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonEmptyAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### NonEmptyIssue

Non empty issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `NonEmptyIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />

### NonNullable

Extracts `null` from a type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `NonNullable` <Property {...properties.NonNullable} />

### NonNullableIssue

Non nullable issue interface.

#### Definition

- `NonNullableIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NonNullableSchema

Non nullable schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonNullableSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NonNullableSchemaAsync

Non nullable schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonNullableSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NonNullish

Extracts `null` and `undefined` from a type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `NonNullish` <Property {...properties.NonNullish} />

### NonNullishIssue

Non nullish issue interface.

#### Definition

- `NonNullishIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NonNullishSchema

Non nullish schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonNullishSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NonNullishSchemaAsync

Non nullish schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonNullishSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NonOptional

Extracts `undefined` from a type.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `NonOptional` <Property {...properties.NonOptional} />

### NonOptionalIssue

Non optional issue interface.

#### Definition

- `NonOptionalIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NonOptionalSchema

Non optional schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonOptionalSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NonOptionalSchemaAsync

Non optional schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NonOptionalSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `message` <Property {...properties.message} />

### NormalizeAction

Normalize action interface.

#### Generics

- `TForm` <Property {...properties.TForm} />

#### Definition

- `NormalizeAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `form` <Property {...properties.form} />

### NormalizeForm

Normalize form type.

#### Definition

- `NormalizeForm` <Property {...properties.NormalizeForm} />

### NotBytesAction

Not bytes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotBytesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotBytesIssue

Not bytes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotBytesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NotEntriesAction

Not entries action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotEntriesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotEntriesIssue

Not entries issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotEntriesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NotGraphemesAction

Not graphemes action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotGraphemesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotGraphemesIssue

Not graphemes issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotGraphemesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NotLengthAction

Not length action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotLengthAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotLengthIssue

Not length issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotLengthIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NotSizeAction

Not size action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotSizeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotSizeIssue

Not size issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotSizeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NotValueAction

Not value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotValuesAction

Not values action type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotValuesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotValueIssue

Not value issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### NotValuesIssue

Not values issue type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotValuesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### NotWordsAction

Not words action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NotWordsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `locales` <Property {...properties.locales} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### NotWordsIssue

Not words issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `NotWordsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### NullableSchema

Nullable schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `NullableSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### NullableSchemaAsync

Nullable schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `NullableSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### NullishSchema

Nullish schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `Nullish` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### NullishSchemaAsync

Nullish schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `Nullish` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### NullIssue

Null issue interface.

#### Definition

- `NullIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NullSchema

Null schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NullSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### NumberIssue

Number issue interface.

#### Definition

- `NumberIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### NumberSchema

Number schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `NumberSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### ObjectEntries

Object entries interface.

#### Definition

- `ObjectEntries` <Property {...properties.ObjectEntries} />

### ObjectEntriesAsync

Object entries async interface.

#### Definition

- `ObjectEntriesAsync` <Property {...properties.ObjectEntriesAsync} />

### ObjectIssue

Object issue interface.

#### Definition

- `ObjectIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### ObjectKeys

Object keys type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `ObjectKeys` <Property {...properties.ObjectKeys} />

### ObjectPathItem

Object path item interface.

#### Definition

- `ObjectPathItem`
  - `type` <Property {...properties.type} />
  - `origin` <Property {...properties.origin} />
  - `input` <Property {...properties.input} />
  - `key` <Property type="string" />
  - `value` <Property type="unknown" />

The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

### ObjectSchema

Object schema interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ObjectSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### ObjectSchemaAsync

Object schema async interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ObjectSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### ObjectWithRestIssue

Object with rest issue interface.

#### Definition

- `ObjectWithRestIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### ObjectWithRestSchema

Object with rest schema interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ObjectWithRestSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `rest` <Property {...properties.rest} />
  - `message` <Property {...properties.message} />

### ObjectWithRestSchemaAsync

Object schema async interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ObjectWithRestSchemaAsync` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `rest` <Property {...properties.rest} />
  - `message` <Property {...properties.message} />

### OctalAction

Octal action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `OctalAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### OctalIssue

Octal issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `OctalIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### OptionalSchema

Optional schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `OptionalSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### OptionalSchemaAsync

Optional schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `OptionalSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### OutputDataset

Output dataset interface.

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `OutputDataset` <Property {...properties.OutputDataset} />

### ParseJsonAction

JSON parse action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TConfig` <Property {...properties.TConfig} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ParseJsonAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `config` <Property {...properties.config} />
  - `message` <Property {...properties.message} />

### ParseJsonConfig

JSON parse config interface.

#### Definition

- `ParseJsonConfig`
  - `reviver` <Property {...properties.reviver} />

### ParseJsonIssue

JSON parse issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ParseJsonIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />

### Parser

The parser interface.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Definition

- `Parser`
  - <Property {...properties.function} />
  - `schema` <Property {...properties.schema} />
  - `config` <Property {...properties.config} />

### ParserAsync

The parser async interface.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Definition

- `ParserAsync`
  - <Property {...properties.function} />
  - `schema` <Property {...properties.schema} />
  - `config` <Property {...properties.config} />

### PartialCheckAction

Partial check action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TPaths` <Property {...properties.TPaths} />
- `TSelection` <Property {...properties.TSelection} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `PartialCheckAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `paths` <Property {...properties.paths} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### PartialCheckActionAsync

Partial check action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TPaths` <Property {...properties.TPaths} />
- `TSelection` <Property {...properties.TSelection} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `PartialCheckActionAsync` <Property {...properties.BaseValidationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `paths` <Property {...properties.paths} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### PartialCheckIssue

Partial check issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `PartialCheckIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### PartialDataset

Partial dataset interface.

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `UntypedDataset`
  - `typed` <Property {...properties.typed} />
  - `value` <Property {...properties.value} />
  - `issues` <Property {...properties.issues} />

### PartialInput

Partial input type.

#### Definition

- `PartialInput` <Property {...properties.PartialInput} />

### Path

Path type.

#### Definition

- `Path` <Property {...properties.Path} />

### PicklistOptions

Picklist options type.

#### Definition

- `PicklistOptions` <Property {...properties.PicklistOptions} />

### PicklistIssue

Picklist issue interface.

#### Definition

- `PicklistIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### PicklistSchema

Picklist schema interface.

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `PicklistSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### PipeAction

Pipe action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `PipeAction` <Property {...properties.PipeAction} />

### PipeActionAsync

Pipe action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `PipeActionAsync` <Property {...properties.PipeActionAsync} />

### PipeItem

Pipe item type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `PipeItem` <Property {...properties.PipeItem} />

### PipeItemAsync

Pipe item async type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />
- `TIssue` <Property {...properties.TIssue} />

#### Definition

- `PipeItemAsync` <Property {...properties.PipeItemAsync} />

### PromiseIssue

Promise issue interface.

#### Definition

- `PromiseIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### PromiseSchema

Promise schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `PromiseSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### RawCheckAction

Raw check action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `RawCheckAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### RawCheckActionAsync

Raw check action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `RawCheckActionAsync` <Property {...properties.BaseValidationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />

### RawCheckIssue

Raw check issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `RawCheckIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />

### RawTransformAction

Raw transform action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `RawTransformAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### RawTransformActionAsync

Raw transform action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `RawTransformActionAsync` <Property {...properties.BaseTransformationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### RawTransformIssue

Raw transform issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `RawTransformIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />

### ReadonlyAction

Readonly action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `ReadonlyAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### RecordIssue

Record issue interface.

#### Definition

- `RecordIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### RecordSchema

Record schema interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `RecordSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### RecordSchemaAsync

Record schema async interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `RecordSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### ReduceItemsAction

Reduce items action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `ReduceItemsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />
  - `initial` <Property {...properties.initial} />

### RegexAction

Regex action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `RegexAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### RegexIssue

Regex issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `RegexIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### RequiredPath

Required path type.

#### Definition

- `RequiredPath` <Property {...properties.RequiredPath} />

### RequiredPaths

Required paths type.

#### Definition

- `RequiredPaths` <Property {...properties.RequiredPaths} />

### ReturnsAction

Returns action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `ReturnsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `schema` <Property {...properties.schema} />

### ReturnsActionAsync

Returns action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `ReturnsActionAsync` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `schema` <Property {...properties.schema} />

### RfcEmailAction

RFC email action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `EmailAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### RfcEmailIssue

RFC email issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `EmailIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### SafeIntegerAction

Safe integer action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SafeIntegerAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### SafeIntegerIssue

Safe integer issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `SafeIntegerIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### SafeParser

The safe parser interface.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Definition

- `SafeParser`
  - <Property {...properties.function} />
  - `schema` <Property {...properties.schema} />
  - `config` <Property {...properties.config} />

### SafeParserAsync

The safe parser async interface.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TConfig` <Property {...properties.TConfig} />

#### Definition

- `SafeParserAsync`
  - <Property {...properties.function} />
  - `schema` <Property {...properties.schema} />
  - `config` <Property {...properties.config} />

### SafeParseResult

Safe parse result type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `SafeParseResult`
  - `typed` <Property {...properties.typed} />
  - `success` <Property {...properties.success} />
  - `output` <Property {...properties.output} />
  - `issues` <Property {...properties.issues} />

### SchemaWithFallback

Schema with fallback type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TFallback` <Property {...properties.TFallback} />

#### Definition

- `SchemaWithFallback` <Property {...properties.BaseSchema} />
  - `fallback` <Property {...properties.fallback} />

### SchemaWithFallbackAsync

Schema with fallback async type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />
- `TFallback` <Property {...properties.TFallback} />

#### Definition

- `SchemaWithFallbackAsync` <Property {...properties.ModifiedBaseSchemaOrAsync} />
  - `fallback` <Property {...properties.fallback} />
  - `async` <Property {...properties.async} />
  - `~run` <Property {...properties['~run']} />

### SchemaWithoutPipe

Schema without pipe type.

#### Generics

- `TSchema` <Property {...properties.TSchema} />

#### Definition

- `SchemaWithoutPipe` <Property {...properties.SchemaWithoutPipe} />

### SchemaWithPartial

Schema with partial type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/partial/partial.ts).

### SchemaWithPartialAsync

Schema with partial async type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/partial/partialAsync.ts).

### SchemaWithPipe

Schema with pipe type.

#### Generics

- `TPipe` <Property {...properties.TPipe} />

#### Definition

- `SchemaWithPipe` <Property {...properties.SchemaWithPipe} />
  - `pipe` <Property {...properties.pipe} />
  - `~types` <Property {...properties['~types']} />
  - `~run` <Property {...properties['~run']} />

### SchemaWithPipeAsync

Schema with pipe async type.

#### Generics

- `TPipe` <Property {...properties.TPipe} />

#### Definition

- `SchemaWithPipeAsync` <Property {...properties.SchemaWithPipe} />
  - `pipe` <Property {...properties.pipe} />
  - `async` <Property {...properties.async} />
  - `~types` <Property {...properties['~types']} />
  - `~run` <Property {...properties['~run']} />

### SchemaWithRequired

Schema with required type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/required/required.ts).

### SchemaWithRequiredAsync

Schema with required async type.

> This type is too complex to display. Please refer to the [source code](https://github.com/fabian-hiller/valibot/blob/main/library/src/methods/required/requiredAsync.ts).

### SetPathItem

Set path item interface.

#### Definition

- `SetPathItem` <Property type="object" />
  - `type` <Property {...properties.type} />
  - `origin` <Property {...properties.origin} />
  - `input` <Property {...properties.input} />
  - `value` <Property type="unknown" />

The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

### RecordIssue

Record issue interface.

#### Definition

- `RecordIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### SetSchema

Set schema interface.

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SetSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### SetSchemaAsync

Set schema async interface.

#### Generics

- `TValue` <Property {...properties.TValue} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SetSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `value` <Property {...properties.value} />
  - `message` <Property {...properties.message} />

### SizeAction

Size action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SizeAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### SizeInput

Size input type.

#### Definition

- `SizeInput` <Property {...properties.SizeInput} />

### SizeIssue

Size issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `SizeIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### SlugAction

Slug action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SlugAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### SlugIssue

Slug issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `SlugIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### SomeItemAction

Some action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SomeItemAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### SomeItemIssue

Some item issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `SomeItemIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### SortItemsAction

Sort items action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `SortItemsAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### StandardFailureResult

The result interface if validation fails.

#### Definition

- `StandardFailureResult`
  - `issues` <Property {...properties.issues} />

### StandardIssue

The issue interface of the failure output.

#### Definition

- `StandardIssue`
  - `message` <Property {...properties.message} />
  - `path` <Property {...properties.path} />

### StandardPathItem

The path item interface of the issue.

#### Definition

- `StandardPathItem`
  - `key` <Property {...properties.key} />

### StandardProps

The Standard Schema properties interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `StandardProps`
  - `version` <Property {...properties.version} />
  - `vendor` <Property {...properties.vendor} />
  - `validate` <Property {...properties.validate} />
  - `types` <Property {...properties.types} />

### StandardResult

The result interface of the validate function.

#### Generics

- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `StandardResult` <Property {...properties.StandardResult} />

### StandardSuccessResult

The result interface if validation succeeds.

#### Generics

- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `StandardSuccessResult`
  - `value` <Property {...properties.value} />
  - `issues` <Property {...properties.issues} />

### StandardTypes

The Standard Schema types interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `StandardTypes`
  - `input` <Property {...properties.input} />
  - `output` <Property {...properties.output} />

### StartsWithAction

Starts with action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StartsWithAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### StartsWithIssue

Starts with issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `StartsWithIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### StrictObjectIssue

Strict object issue interface.

#### Definition

- `StrictObjectIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### StrictObjectSchema

Strict object schema interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StrictObjectSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### StrictObjectSchemaAsync

Strict object schema async interface.

#### Generics

- `TEntries` <Property {...properties.TEntries} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StrictObjectSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `entries` <Property {...properties.entries} />
  - `message` <Property {...properties.message} />

### StrictTupleIssue

Strict tuple issue interface.

#### Definition

- `StrictTupleIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### StrictTupleSchema

Strict tuple schema interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StrictTupleSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### StrictTupleSchemaAsync

Strict tuple schema async interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StrictTupleSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### RecordIssue

Record issue interface.

#### Definition

- `RecordIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### StringSchema

String schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StringSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### StringifyJsonAction

JSON stringify action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TConfig` <Property {...properties.TConfig} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `StringifyJsonAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `config` <Property {...properties.config} />
  - `message` <Property {...properties.message} />

### StringifyJsonConfig

JSON stringify config interface.

#### Definition

- `StringifyJsonConfig`
  - `replacer` <Property {...properties.replacer} />
  - `space` <Property {...properties.space} />

### StringifyJsonIssue

JSON stringify issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `StringifyJsonIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />

### SuccessDataset

Success dataset interface.

#### Generics

- `TValue` <Property {...properties.TValue} />

#### Definition

- `TypedDataset`
  - `typed` <Property {...properties.typed} />
  - `value` <Property {...properties.value} />
  - `issues` <Property {...properties.issues} />

### SymbolIssue

Symbol issue interface.

#### Definition

- `SymbolIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### SymbolSchema

Symbol schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `SymbolSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### TitleAction

Title action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TTitle` <Property {...properties.TTitle} />

#### Definition

- `TitleAction` <Property {...properties.BaseMetadata} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `title` <Property {...properties.title} />

### ToLowerCaseAction

To lower case action interface.

#### Definition

- `ToLowerCaseAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### ToMinValueAction

To min value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `ToMinValueAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `requirement` <Property {...properties.requirement} />

### ToMaxValueAction

To max value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `ToMaxValueAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `requirement` <Property {...properties.requirement} />

### ToUpperCaseAction

To upper case action interface.

#### Definition

- `ToUpperCaseAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### TransformAction

Transform action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `TransformAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### TransformActionAsync

Transform action async interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TOutput` <Property {...properties.TOutput} />

#### Definition

- `TransformActionAsync` <Property {...properties.BaseTransformationAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `operation` <Property {...properties.operation} />

### TrimAction

Trim action interface.

#### Definition

- `TrimAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### TrimEndAction

Trim end action interface.

#### Definition

- `TrimEndAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### TrimStartAction

Trim start action interface.

#### Definition

- `TrimStartAction` <Property {...properties.BaseTransformation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />

### TupleIssue

Tuple issue interface.

#### Definition

- `TupleIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### TupleItems

Tuple items type.

#### Definition

- `TupleItems` <Property {...properties.TupleItems} />

### TupleItemsAsync

Tuple items async type.

#### Definition

- `TupleItemsAsync` <Property {...properties.TupleItemsAsync} />

### TupleSchema

Tuple schema interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `TupleSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### TupleSchemaAsync

Tuple schema async interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `TupleSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `message` <Property {...properties.message} />

### TupleWithRestIssue

Tuple with rest issue interface.

#### Definition

- `TupleWithRestIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### TupleWithRestSchema

Tuple with rest schema interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `TupleWithRestSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `rest` <Property {...properties.rest} />
  - `message` <Property {...properties.message} />

### TupleWithRestSchemaAsync

Tuple with rest schema async interface.

#### Generics

- `TItems` <Property {...properties.TItems} />
- `TRest` <Property {...properties.TRest} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `TupleWithRestSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `items` <Property {...properties.items} />
  - `rest` <Property {...properties.rest} />
  - `message` <Property {...properties.message} />

### UlidAction

ULID action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UlidAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### UlidIssue

ULID issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `UlidIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### UndefinedableSchema

Undefinedable schema interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `UndefinedableSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### UndefinedableSchemaAsync

Undefinedable schema async interface.

#### Generics

- `TWrapped` <Property {...properties.TWrapped} />
- `TDefault` <Property {...properties.TDefault} />

#### Definition

- `UndefinedableSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `wrapped` <Property {...properties.wrapped} />
  - `default` <Property {...properties.default} />

### UndefinedIssue

Undefined issue interface.

#### Definition

- `UndefinedIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### UndefinedSchema

Undefined schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UndefinedSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### UnionOptions

Union options type.

#### Definition

- `UnionOptions` <Property {...properties.UnionOptions} />

### UnionOptionsAsync

Union options async type.

#### Definition

- `UnionOptionsAsync` <Property {...properties.UnionOptionsAsync} />

### UnionIssue

Union issue interface.

#### Generics

- `TSubIssue` <Property {...properties.TSubIssue} />

#### Definition

- `UnionIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `issues` <Property {...properties.issues} />

### UnionSchema

Union schema interface.

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UnionSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### UnionSchemaAsync

Union schema async interface.

#### Generics

- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UnionSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### UnknownDataset

Unknown dataset interface.

#### Definition

- `TypedDataset`
  - `typed` <Property {...properties.typed} />
  - `value` <Property {...properties.value} />
  - `issues` <Property {...properties.issues} />

### UnknownPathItem

Unknown path item interface.

#### Definition

- `UnknownPathItem`
  - `type` <Property {...properties.type} />
  - `origin` <Property {...properties.origin} />
  - `input` <Property type="unknown" />
  - `key` <Property type="unknown" />
  - `value` <Property type="unknown" />

The `input` of a path item may differ from the `input` of its issue. This is because path items are subsequently added by parent schemas and are related to their input. Transformations of child schemas are not taken into account.

### UnknownSchema

Unknown schema interface.

#### Definition

- `UnknownSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />

### UrlAction

URL action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UrlAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### UrlIssue

URL issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `UrlIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### UuidAction

UUID action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `UuidAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### UuidIssue

UUID issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />

#### Definition

- `UuidIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />

### ValueAction

Value action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ValueAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### ValuesAction

Values action type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `ValuesAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### ValueInput

Value input type.

#### Definition

- `ValueInput` <Property {...properties.ValueInput} />

### ValueIssue

Value issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `ValueIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### ValuesIssue

Values issue type.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `ValuesIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `requirement` <Property {...properties.requirement} />

### VariantIssue

Variant issue interface.

#### Definition

- `VariantIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### VariantOptions

Variant options type.

#### Generics

- `TKey` <Property {...properties.TKey} />

#### Definition

- `VariantOptions` <Property {...properties.VariantOptions} />

### VariantOptionsAsync

Variant options async type.

#### Generics

- `TKey` <Property {...properties.TKey} />

#### Definition

- `VariantOptionsAsync` <Property {...properties.VariantOptionsAsync} />

### VariantSchema

Variant schema interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `VariantSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### VariantSchemaAsync

Variant schema async interface.

#### Generics

- `TKey` <Property {...properties.TKey} />
- `TOptions` <Property {...properties.TOptions} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `VariantSchemaAsync` <Property {...properties.BaseSchemaAsync} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `key` <Property {...properties.key} />
  - `options` <Property {...properties.options} />
  - `message` <Property {...properties.message} />

### VoidIssue

Void issue interface.

#### Definition

- `VoidIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />

### VoidSchema

Void schema interface.

#### Generics

- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `VoidSchema` <Property {...properties.BaseSchema} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `message` <Property {...properties.message} />

### WordsAction

Words action interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TLocales` <Property {...properties.TLocales} />
- `TRequirement` <Property {...properties.TRequirement} />
- `TMessage` <Property {...properties.TMessage} />

#### Definition

- `WordsAction` <Property {...properties.BaseValidation} />
  - `type` <Property {...properties.type} />
  - `reference` <Property {...properties.reference} />
  - `expects` <Property {...properties.expects} />
  - `locales` <Property {...properties.locales} />
  - `requirement` <Property {...properties.requirement} />
  - `message` <Property {...properties.message} />

### WordsIssue

Words issue interface.

#### Generics

- `TInput` <Property {...properties.TInput} />
- `TRequirement` <Property {...properties.TRequirement} />

#### Definition

- `WordsIssue` <Property {...properties.BaseIssue} />
  - `kind` <Property {...properties.kind} />
  - `type` <Property {...properties.type} />
  - `expected` <Property {...properties.expected} />
  - `received` <Property {...properties.received} />
  - `requirement` <Property {...properties.requirement} />
