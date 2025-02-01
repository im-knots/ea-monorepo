import pg from 'pg';
const { Client } = pg;

const db = new Client({
  connectionString: process.env.POSTGRES_URL
});
await db.connect();

export  { db } ;

