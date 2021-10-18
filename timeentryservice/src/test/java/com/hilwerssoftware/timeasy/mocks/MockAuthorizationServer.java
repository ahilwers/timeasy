package com.hilwerssoftware.timeasy.mocks;

import java.util.HashMap;
import java.util.Map;
import java.util.Scanner;

import javax.ws.rs.core.Response;

import com.github.tomakehurst.wiremock.WireMockServer;
import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.jwk.KeyUse;
import com.nimbusds.jose.jwk.RSAKey;
import com.nimbusds.jose.jwk.gen.RSAKeyGenerator;
import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;
import io.restassured.RestAssured;
import io.restassured.response.ResponseBody;

public class MockAuthorizationServer implements QuarkusTestResourceLifecycleManager {
    private WireMockServer wireMockServer;
    public static RSAKey keyPair;

    static {
        try {
            keyPair = new RSAKeyGenerator(2048)
                    .keyID("123")
                    .keyUse(KeyUse.SIGNATURE)
                    .generate();
        } catch (JOSEException e) {
            e.printStackTrace();
        }
    }

    @Override
    public Map<String, String> start() {
        wireMockServer = new WireMockServer(8090);
        wireMockServer.start();

        postStubMapping(oidcConfigurationStub());
        postStubMapping(publicKeysStub(keyPair.toPublicJWK().toJSONString()));

        Map<String,String> props = new HashMap<>();
        props.put("quarkus.oidc.auth-server-url", wireMockServer.baseUrl() + "/mock-server");
        props.put("wiremock.url", wireMockServer.baseUrl());
        return props;
    }

    @Override
    public void stop() {
        if (wireMockServer != null) {
            wireMockServer.stop();
        }
    }

    private ResponseBody<?> postStubMapping(String request) {
        RestAssured.baseURI = wireMockServer.baseUrl();
        return RestAssured.given()
                .body(request)
                .post("/__admin/mappings")
                .then()
                .statusCode(Response.Status.CREATED.getStatusCode())
                .extract()
                .response()
                .body();
    }

    private String oidcConfigurationStub() {
        return readFile("/oidcconfig.json")
                .replace("$baseUrl", wireMockServer.baseUrl());
    }

    private String publicKeysStub(String keys) {
        return readFile("/publickey.json")
                .replace("$keys", keys);
    }

    private String readFile(String fileName) {
        return new Scanner(getClass()
                .getResourceAsStream(fileName), "UTF-8")
                .useDelimiter("\\A")
                .next();
    }
}