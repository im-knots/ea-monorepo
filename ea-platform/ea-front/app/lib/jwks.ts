import { readFileSync } from "fs"
import * as jose from "jose"
import path from "path";


export async function getJWKS() {
  const jwksPath = path.resolve(process.cwd(), "jwks/jwks.json");
  const jwksFile = readFileSync(jwksPath, "utf8");
  return jose.createLocalJWKSet(JSON.parse(jwksFile));
}

export async function getJWKSFile() {
  const jwksPath = path.resolve(process.cwd(), "jwks/jwks.json");
  const jwksFile = readFileSync(jwksPath, "utf8");
  return JSON.parse(jwksFile);

}
export async function getPrivateKey() {
  const privateKeyPath = path.resolve(process.cwd(), "jwks/private.json");
  const privateKeyFile = readFileSync(privateKeyPath, "utf8");
  return jose.importJWK(JSON.parse(privateKeyFile));
}