import { MongoClient } from 'mongodb' 
 
let client 
let clientPromise: Promise<MongoClient> 

function getClientPromise() {  
 if (!process.env.MONGO_URI) { 
     throw new Error('Invalid/Missing      environment variable: "MONGO_URI"') 
  } 

 const uri = process.env.MONGO_URI 
 const options = {} 

 if (process.env.NODE_ENV === 'development') { 
   // In development mode, use a global variable so that the value 
   // is preserved across module reloads caused by HMR (Hot Module Replacement). 
   let globalWithMongo = global as typeof globalThis & { 
     _mongoClientPromise?: Promise<MongoClient> 
   } 

   if (!globalWithMongo._mongoClientPromise) { 
     client = new MongoClient(uri, options) 
     globalWithMongo._mongoClientPromise = client.connect() 
   } 
   clientPromise = globalWithMongo._mongoClientPromise 
 } else { 
   // In production mode, it's best to not use a global variable. 
   client = new MongoClient(uri, options) 
   clientPromise = client.connect() 
 } 
 return clientPromise
}
  
export default getClientPromise