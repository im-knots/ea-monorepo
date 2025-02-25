import { NextResponse } from 'next/server';
import * as k8s from '@kubernetes/client-node';

const KUBERNETES_NAMESPACE = process.env.KUBERNETES_NAMESPACE || 'ea-platform';

// Load Kubernetes Configuration
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sApi = k8s.KubernetesObjectApi.makeApiClient(kc);

// Function to delete a Kubernetes object
async function deleteK8sObject(kind: string, name: string) {
  try {
    console.log(`🗑️ Attempting to delete ${kind}: ${name}`);
    await k8sApi.delete({
      apiVersion: kind === "Secret" ? "v1" : "configuration.konghq.com/v1",
      kind,
      metadata: { name, namespace: KUBERNETES_NAMESPACE },
    });
    console.log(`✅ Successfully deleted ${kind}: ${name}`);
  } catch (error: any) {
    if (error.body?.reason === "NotFound") {
      console.log(`ℹ️ ${kind} ${name} not found, nothing to delete.`);
    } else {
      console.error(`❌ Failed to delete ${kind} ${name}:`, error);
    }
  }
}

export async function POST(req: Request) {
  try {
    const token = req.headers.get('cookie')?.split('token=')[1]?.split(';')[0];

    if (!token) {
      return NextResponse.json({ message: 'No active session' }, { status: 200 });
    }

    // Extract user ID from JWT
    let userId: string | undefined;
    try {
      const decoded = JSON.parse(atob(token.split('.')[1]));
      userId = decoded?.userId;
    } catch (error) {
      console.error('Invalid token:', error);
      return NextResponse.json({ error: 'Invalid token' }, { status: 400 });
    }

    if (!userId) {
      return NextResponse.json({ error: 'User ID missing in token' }, { status: 400 });
    }

    console.log(`🔹 Logging out user: ${userId}`);

    // Delete K8s Secret and KongConsumer
    await deleteK8sObject("Secret", `kong-jwt-${userId}`);
    await deleteK8sObject("KongConsumer", `kong-consumer-${userId}`);

    // Remove the JWT token cookie
    const response = NextResponse.json({ message: 'Logged out successfully' }, { status: 200 });
    response.cookies.set('token', '', {
      httpOnly: true,
      expires: new Date(0), // Expire the cookie
      path: '/',
    });

    return response;
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json({ error: 'Server error' }, { status: 500 });
  }
}
