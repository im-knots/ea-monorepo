import bcrypt from "bcryptjs";
import { v4 as uuidv4 } from "uuid";
import mongoClient from "./mongodb";
import * as jose from "jose";
import { cookies } from "next/headers";
import { getJWKS, getPrivateKey } from "./jwks";
import * as k8s from '@kubernetes/client-node';

const ALPHA_ACCESS_CODE = process.env.ALPHA_CODE || 'some-string';
const KUBERNETES_NAMESPACE = process.env.KUBERNETES_NAMESPACE || 'ea-platform';

// Load Kubernetes Configuration
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sApi = k8s.KubernetesObjectApi.makeApiClient(kc);


export async function Register(user: { email: string; password: string; alphaCode: string }) {
  const { email, password, alphaCode } = user;
  const hashedPassword = await bcrypt.hash(password, 10);
  const mongodb = await mongoClient();

  if (alphaCode !== ALPHA_ACCESS_CODE) {
    throw new Error("Invalid Alpha Access Code")
  }

  const userRecord = await mongodb.db().collection("users").findOne({ email });

  if (userRecord) {
    console.log("User already exists");
    throw new Error("User already exists");
  }

  const userId = uuidv4();

  // Kubernetes resource names
  const secretName = `third-party-user-creds-${userId}`;
  const serviceAccountName = `sa-user-${userId}`;
  const roleName = `role-user-${userId}`;
  const roleBindingName = `rb-user-${userId}`;

  // Create Kubernetes resources for the user
  await createK8sSecret(secretName);
  await createK8sServiceAccount(serviceAccountName);
  await createK8sRole(roleName, secretName);
  await createK8sRoleBinding(roleBindingName, roleName, serviceAccountName);

  await mongodb.db().collection("users").insertOne({
    id: userId,
    email,
    password: hashedPassword,
  });
}

export async function Login(user: { email: string, password: string }) {
  const cookieStore = await cookies();

  const { email, password } = user;

  const mongodb = await mongoClient();
  const userRecord = await mongodb.db().collection("users").findOne({ email });

  if (!userRecord) {
    console.log("User not found");
    throw new Error("User not found");
  }

  const passwordMatch = await bcrypt.compare(password, userRecord.password);

  if (!passwordMatch) {
    console.log("Invalid password");
    throw new Error("Invalid password");
  }
  try {
    const privateKey = await getPrivateKey();
    const token = await new jose.SignJWT({
      email: userRecord.email,
      iss: "eru-labs-jwt-issuer",
      sub: userRecord.id,
    })
    .setProtectedHeader({ alg: "RS256" })
    .setIssuedAt()
    .setExpirationTime("6h")
    .sign(privateKey);
  
    cookieStore.set('token', token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
    })
    return token;
  } catch (error) {
    console.log("Error creating token", error);
    throw new Error("Error creating token");
  }
}

export async function GetTokenClaims(token: string) {
  if (!token) {
    return {};
  }
  try {
    const JWKS = await getJWKS();
    const { payload } = await jose.jwtVerify(token, JWKS);
    return payload;
  } catch (error) {
    console.log("Invalid token", error);
    return {};
  }
}

// Function to create Kubernetes Secret for 3rd party api cred storage
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
      dummy: "dummy",  // Placeholder to initialize the secret
    },
  };

  try {
    await k8sApi.create(k8sSecret);
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      return;
    }
    throw error;
  }
}

// Function to create Kubernetes Service Account for the User
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
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      return;
    }
    throw error;
  }
}

// Function to create RBAC Role for the user service account
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
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      return;
    }
    throw error;
  }
}

// Function to create RoleBinding for the user service account to its RBAC
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
  } catch (error: any) {
    if (error.body?.reason === 'AlreadyExists') {
      return;
    }
    throw error;
  }
}