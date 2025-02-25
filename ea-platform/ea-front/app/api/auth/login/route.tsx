import { NextResponse } from 'next/server';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { connectToDatabase } from '@/lib/mongodb';
import * as k8s from '@kubernetes/client-node';

const JWT_SECRET = process.env.JWT_SECRET || 'supersecretkey';
const KUBERNETES_NAMESPACE = process.env.KUBERNETES_NAMESPACE || 'ea-platform';

// Load Kubernetes Configuration
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sApi = k8s.KubernetesObjectApi.makeApiClient(kc);

// Function to create or update a Kubernetes object
async function createK8sObject(object: any) {
  try {
    const response = await k8sApi.create(object);
    return response.body;
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      console.log(`Updating existing resource: ${object.metadata.name}`);
      return await k8sApi.patch(object);
    }
    throw error;
  }
}

export async function POST(req: Request) {
    try {
      const { email, password } = await req.json();
  
      if (!email || !password) {
        return NextResponse.json({ error: 'Email and password required' }, { status: 400 });
      }
  
      const { db } = await connectToDatabase();
      const user = await db.collection('users').findOne({ email });
  
      if (!user) {
        return NextResponse.json({ error: 'Invalid credentials' }, { status: 401 });
      }
  
      const passwordMatch = await bcrypt.compare(password, user.password);
      if (!passwordMatch) {
        return NextResponse.json({ error: 'Invalid credentials' }, { status: 401 });
      }
  
      // 🔹 Generate JWT with UUID as 'iss'
      const token = jwt.sign(
        {
          userId: user.id,
          email,
          iss: user.id,  // ✅ Explicitly set issuer to user.id
        },
        JWT_SECRET,
        { expiresIn: '6h' }
      );
  
      // Store token in an `httpOnly` secure cookie
      const response = NextResponse.json({ message: 'Login successful' }, { status: 200 });
      response.cookies.set('token', token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 6 * 60 * 60, // 6 hours
        path: '/',
      });
  
      const k8sSecret = {
        apiVersion: 'v1',
        kind: 'Secret',
        metadata: {
          name: `kong-jwt-${user.id}`,
          namespace: KUBERNETES_NAMESPACE,
          labels: {
            "konghq.com/credential": "jwt",  // ✅ Ensure Kong recognizes it as a JWT credential
          },
          annotations: {
            "konghq.com/consumer": `kong-consumer-${user.id}`,
          },
        },
        type: 'Opaque',
        stringData: {
          algorithm: "HS256",  // ✅ Required by Kong
          key: user.id,        // ✅ This must match the `iss` field in the JWT
          secret: JWT_SECRET,  // ✅ Used for verifying JWTs in Kong
        },
      };
      
  
      await createK8sObject(k8sSecret);
  
      // 🔹 Step 2: Create KongConsumer with UUID as username
      const kongConsumer = {
        apiVersion: 'configuration.konghq.com/v1',
        kind: 'KongConsumer',
        metadata: {
          name: `kong-consumer-${user.id}`,
          namespace: KUBERNETES_NAMESPACE,
          annotations: {
            "kubernetes.io/ingress.class": "kong",
          },
        },
        username: user.id,  // ✅ UUID as username (matches JWT "iss")
        custom_id: email,    // ✅ Email as custom_id
        credentials: [`kong-jwt-${user.id}`],  // ✅ Attach the JWT credential
      };
      
      await createK8sObject(kongConsumer);
  
      return response;
    } catch (error) {
      console.error('Login error:', error);
      return NextResponse.json({ error: 'Server error' }, { status: 500 });
    }
  }
  
