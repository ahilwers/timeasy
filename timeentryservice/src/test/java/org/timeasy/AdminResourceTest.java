package org.timeasy;

import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.security.TestSecurity;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;

@QuarkusTest
public class AdminResourceTest {

    @Test
    @TestSecurity(user="testUser", roles = {"admin"})
    public void TestAdminEndpoint() {
        given()
                .contentType("application/json")
                .get("/api/admin")
                .then()
                .statusCode(200)
                .body(is("granted"));

    }
}
