import { NextResponse } from 'next/server';
import bcrypt from 'bcryptjs';
import { v4 as uuidv4 } from 'uuid';
import { connectToDatabase } from '@/app/lib/mongodb';
import * as k8s from '@kubernetes/client-node';

const ALPHA_ACCESS_CODE = 'some-alpha-string';
const KUBERNETES_NAMESPACE = process.env.KUBERNETES_NAMESPACE || 'ea-platform';

// 🔹 Load Kubernetes Configuration
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sApi = k8s.KubernetesObjectApi.makeApiClient(kc);

// 🔹 Function to create Kubernetes Secret
async function createK8sSecret(secretName: string) {
  const k8sSecret = {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: secretName,
      namespace: KUBERNETES_NAMESPACE,
    },
    type: 'Opaque',
    stringData: {
      dummy: "dummy",  // 🔹 Placeholder to initialize the secret
    },
  };

  try {
    await k8sApi.create(k8sSecret);
    console.log(`✅ Created Kubernetes Secret: ${secretName}`);
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      console.log(`⚠️ Secret ${secretName} already exists. Skipping creation.`);
      return;
    }
    throw error;
  }
}

// 🔹 Function to create Kubernetes Service Account
async function createK8sServiceAccount(serviceAccountName: string) {
  const serviceAccount = {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: serviceAccountName,
      namespace: KUBERNETES_NAMESPACE,
    },
  };

  try {
    await k8sApi.create(serviceAccount);
    console.log(`✅ Created Kubernetes Service Account: ${serviceAccountName}`);
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      console.log(`⚠️ Service Account ${serviceAccountName} already exists. Skipping creation.`);
      return;
    }
    throw error;
  }
}

// 🔹 Function to create RBAC Role for the user
async function createK8sRole(roleName: string, secretName: string) {
  const role = {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'Role',
    metadata: {
      name: roleName,
      namespace: KUBERNETES_NAMESPACE,
    },
    rules: [
      {
        apiGroups: [""], // Core API group
        resources: ["secrets"],
        resourceNames: [secretName], // Restrict access to only this user's secret
        verbs: ["get", "list"], // Read-only permissions
      },
      {
        apiGroups: [""],
        resources: ["events"],
        verbs: ["create", "patch", "update"],
      },
    ],
  };

  try {
    await k8sApi.create(role);
    console.log(`✅ Created Kubernetes Role: ${roleName}`);
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      console.log(`⚠️ Role ${roleName} already exists. Skipping creation.`);
      return;
    }
    throw error;
  }
}

// 🔹 Function to create RoleBinding for the user
async function createK8sRoleBinding(roleBindingName: string, roleName: string, serviceAccountName: string) {
  const roleBinding = {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'RoleBinding',
    metadata: {
      name: roleBindingName,
      namespace: KUBERNETES_NAMESPACE,
    },
    subjects: [
      {
        kind: "ServiceAccount",
        name: serviceAccountName,
        namespace: KUBERNETES_NAMESPACE,
      },
    ],
    roleRef: {
      kind: "Role",
      name: roleName,
      apiGroup: "rbac.authorization.k8s.io",
    },
  };

  try {
    await k8sApi.create(roleBinding);
    console.log(`✅ Created Kubernetes RoleBinding: ${roleBindingName}`);
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      console.log(`⚠️ RoleBinding ${roleBindingName} already exists. Skipping creation.`);
      return;
    }
    throw error;
  }
}

export async function POST(req: Request) {
  try {
    const { name, email, password, alphaCode } = await req.json();

    if (!name || !email || !password || !alphaCode) {
      return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
    }

    if (alphaCode !== ALPHA_ACCESS_CODE) {
      return NextResponse.json({ error: 'Invalid Alpha Access Code' }, { status: 403 });
    }

    const { db } = await connectToDatabase();
    const existingUser = await db.collection('users').findOne({ email });

    if (existingUser) {
      return NextResponse.json({ error: 'User already exists' }, { status: 400 });
    }

    const hashedPassword = await bcrypt.hash(password, 10);
    const userId = uuidv4();

    const newUser = {
      id: userId,
      name,
      email,
      password: hashedPassword,
      compute_credits: 0,
      compute_devices: [],
      jobs: [],
      created_time: new Date(),
    };

    await db.collection('users').insertOne(newUser);

    // 🔹 Kubernetes resource names
    const secretName = `third-party-user-creds-${userId}`;
    const serviceAccountName = `sa-user-${userId}`;
    const roleName = `role-user-${userId}`;
    const roleBindingName = `rb-user-${userId}`;

    // 🔹 Create Kubernetes resources for the user
    await createK8sSecret(secretName);
    await createK8sServiceAccount(serviceAccountName);
    await createK8sRole(roleName, secretName);
    await createK8sRoleBinding(roleBindingName, roleName, serviceAccountName);

    return NextResponse.json({ 
      message: 'User created successfully', 
      id: userId, 
      name: newUser.name, 
      created_time: newUser.created_time 
    }, { status: 201 });

  } catch (error) {
    console.error('❌ Error in /register:', error);
    return NextResponse.json({ error: 'Server error' }, { status: 500 });
  }
}
