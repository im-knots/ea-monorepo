import { MongoClient } from 'mongodb';

// const MONGO_URI = process.env.MONGO_URI || 'mongodb://mongodb.ea-platform.svc.cluster.local:27017/ainuUsers';
const MONGO_URI = process.env.MONGO_URI || 'mongodb://localhost:8086/ainuUsers';
let cachedClient: MongoClient | null = null;

export async function connectToDatabase() {
  if (!cachedClient) {
    console.log('🔄 Creating new MongoDB connection...');
    cachedClient = new MongoClient(MONGO_URI);
  }

  try {
    // 🔹 Check if the existing connection is still alive
    await cachedClient.db().admin().ping();
  } catch (error) {
    console.error('⚠️ Lost MongoDB connection, reconnecting...', error);
    cachedClient = new MongoClient(MONGO_URI);
    await cachedClient.connect();
  }

  console.log('✅ MongoDB Connected');
  return { db: cachedClient.db() };
}
