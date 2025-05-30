# Ea-Front
A NextJS Front-end client for all things Erulabs.

## Requirements

 - NodeJS v22
 - pnpm

 ## Usage

 Clone the repository, install dependencies, run the test, run the project:

```bash
$ git clone git@github.com:/eru-labs/eru-labs-monorepo && cd eru-labs-monorepo/ea-platform/ea-front
$ pnpm install 
$ pnpx @tailwindcss/cli -i ./app/global.css
$ pnpm test
$ $ pnpm test

> ea-front@1.0.0 test 
> jest
 PASS  __tests__/root.test.tsx
  Page
    ✓ renders the root page (8 ms)
    ✓ renders the login page (24 ms)

Test Suites: 1 passed, 1 total
Tests:       2 passed, 2 total
Snapshots:   0 total
Time:        1.062 s
Ran all test suites.
```

Provide the required environment variables, with examples available in `.env.example` 
```
$ cp .env.example .env
$ cat .env
DATABASE_URL="postgresql://admin:password@localhost:5432/ea?schema=public"
JWT_SECRET=super-secret
```

Run Prisma migrations to prepare your PostgreSQL instance, then run the project:

```
$ npx prisma migrate dev
Environment variables loaded from .env
Prisma schema loaded from prisma/schema.prisma
Datasource "db": PostgreSQL database "ea", schema "public" at "localhost:5432"
...

$ pnpm run dev
> ea-front@1.0.0 dev
> next dev --turbopack

   ▲ Next.js 15.1.6 (Turbopack)
   - Local:        http://localhost:3000
   - Network:      http://x.x.x.x:3000

 ✓ Starting...
 ✓ Ready in 887ms
```
```

## Production build

A Dockerfile and Helm chart are available for production deployment.
