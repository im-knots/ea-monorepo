import { MongoClient } from 'mongodb';

// const MONGO_URI = process.env.MONGO_URI || 'mongodb://mongodb.ea-platform.svc.cluster.local:27017/ainuUsers';
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:8086/ainuUsers';
let cachedClient: MongoClient | null = null;

export async function connectToDatabase() {
  if (!cachedClient) {
    cachedClient = new MongoClient(MONGO_URI);
    await cachedClient.connect();
  }
  const db = cachedClient.db();
  return { db };
}
