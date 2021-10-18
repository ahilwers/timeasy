package com.hilwerssoftware.timeasy;

import com.hilwerssoftware.timeasy.mocks.MockAuthorizationServer;
import io.quarkus.test.common.QuarkusTestResource;
import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;

@QuarkusTest
@QuarkusTestResource(MockAuthorizationServer.class)
public class AdminResourceTest extends BaseResourceTest {

    private static final String BEARER_TOKEN = "337aab0f-b547-489b-9dbd-a54dc7bdf20d";

    @Test
    public void TestAdminEndpoint() {
        given()
                .contentType("application/json")
                .auth()
                .oauth2(generateJWT("user"))
                .get("/api/admin")
                .then()
                .statusCode(200)
                .body(is("granted"));

    }
}
