import * as jose from "jose";
import fs from "fs";

const { publicKey, privateKey } = await jose.generateKeyPair("RS256", {
  extractable: true,
});

const exportedPublicKey = await jose.exportJWK(publicKey)
exportedPublicKey.use = "sig";
exportedPublicKey.kid = Date.now().toString();
exportedPublicKey.alg = "RS256";

const exportedPrivateKey = await jose.exportJWK(privateKey);
exportedPrivateKey.use = "sig";
exportedPrivateKey.kid = Date.now().toString();
exportedPrivateKey.alg = "RS256";

const JWKS = {
  keys: [
    exportedPublicKey
  ]
};

fs.mkdirSync("./jwks", { recursive: true }, (err) => {
  if (err) throw err;
});

fs.writeFileSync("./jwks/jwks.json", JSON.stringify(JWKS, null, 2), (err) => {
  if (err) throw err;
});

fs.writeFileSync("./jwks/private.json", JSON.stringify(exportedPrivateKey, null, 2), (err) => {
  if (err) throw err;
});
