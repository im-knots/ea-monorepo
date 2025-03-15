import { readFileSync } from "fs"
import * as jose from "jose"

const privateKeyFile = readFileSync("jwks/private.json", "utf8");
export const jwksFile = readFileSync("jwks/jwks.json", "utf8");

export const privateKey = await jose.importJWK(JSON.parse(privateKeyFile));
export const JWKS = jose.createLocalJWKSet(JSON.parse(jwksFile));

