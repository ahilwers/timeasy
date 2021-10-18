package com.hilwerssoftware.timeasy;

import com.hilwerssoftware.timeasy.mocks.MockAuthorizationServer;
import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.JOSEObjectType;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.crypto.RSASSASigner;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;

import java.util.Arrays;
import java.util.Date;

public class BaseResourceTest {

    protected String generateJWT(String role) {
        // Prepare JWT with claims set
        SignedJWT signedJWT = new SignedJWT(
                new JWSHeader.Builder(JWSAlgorithm.RS256)
                        .keyID(MockAuthorizationServer.keyPair.getKeyID())
                        .type(JOSEObjectType.JWT)
                        .build(),
                new JWTClaimsSet.Builder()
                        .subject("backend-service")
                        .issuer("https://wiremock")
                        .claim(
                                "realm_access",
                                new JWTClaimsSet.Builder()
                                        .claim("roles", Arrays.asList(role))
                                        .build()
                                        .toJSONObject()
                        )
                        .claim("scope", "openid email profile")
                        .expirationTime(new Date(new Date().getTime() + 60 * 1000))
                        .build()
        );
        // Compute the RSA signature
        try {
            signedJWT.sign(new RSASSASigner(MockAuthorizationServer.keyPair.toRSAKey()));
        } catch (JOSEException e) {
            throw new RuntimeException(e);
        }
        return signedJWT.serialize();
    }

}
