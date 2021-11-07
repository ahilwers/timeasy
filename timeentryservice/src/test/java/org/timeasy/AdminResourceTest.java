package org.timeasy;

import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;

@QuarkusTest
public class AdminResourceTest {

    @InjectMock
    SecurityIdentity securityIdentity;

    @BeforeEach
    public void setup() {
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
    }

    @Test
    public void TestAdminEndpoint() {
        given()
                .contentType("application/json")
                .get("/api/admin")
                .then()
                .statusCode(200)
                .body(is("granted"));

    }
}
