package org.timeasy.resources;

import static io.restassured.RestAssured.given;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;

import java.util.List;

import javax.inject.Inject;
import javax.transaction.Transactional;
import javax.ws.rs.core.MediaType;

import com.github.dockerjava.zerodep.shaded.org.apache.hc.core5.http.HttpStatus;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.services.ProjectService;
import org.timeasy.services.UserDataService;

import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;
import io.quarkus.test.security.TestSecurity;
import io.vertx.core.json.JsonObject;

@QuarkusTest
public class ProjectResourceTest {

    @InjectMock
    SecurityIdentity securityIdentity;
    @InjectMock
    UserDataService userDataService;
    @Inject
    ProjectRepository projectRepository;
    @Inject
    ProjectService projectService;

    @BeforeEach
    @Transactional
    public void setup() {
        projectRepository.deleteAll();
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
        Mockito.when(userDataService.getUserId(any())).thenReturn("user1");
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void canProjectBeAddedViaService() {
        JsonObject jsonObject = new JsonObject()
                .put("description", "Project");
        given()
                .contentType(MediaType.APPLICATION_JSON)
                .body(jsonObject.toString())
                .when()
                .post("/api/v1/projects")
                .then()
                .assertThat()
                .statusCode(HttpStatus.SC_OK);
        List<Project> projects = projectService.listAll();
        assertEquals(1, projects.size());
        Project project = projects.get(0);
        assertEquals("Project", project.getDescription());
        // The project should belong to the correct user:
        assertEquals("user1", project.getUserId());
    }
}
