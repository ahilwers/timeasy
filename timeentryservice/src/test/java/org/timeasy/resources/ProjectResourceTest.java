package org.timeasy.resources;

import static io.restassured.RestAssured.given;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.hamcrest.CoreMatchers.hasItems;
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.CoreMatchers.equalTo;

import java.util.List;

import javax.inject.Inject;
import javax.transaction.Transactional;
import javax.ws.rs.core.MediaType;

import com.github.dockerjava.zerodep.shaded.org.apache.hc.core5.http.HttpStatus;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.services.ProjectService;
import org.timeasy.services.UserDataService;
import org.timeasy.tools.EntityExistsException;

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

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void addingAProjectViaServiceFailsIfItExists() throws EntityExistsException {
        Project project = new Project();
        projectService.add(project);
        JsonObject jsonObject = new JsonObject()
                .put("id", project.getId());
        given()
                .contentType(MediaType.APPLICATION_JSON)
                .body(jsonObject.toString())
                .when()
                .post("/api/v1/projects")
                .then()
                .assertThat()
                .statusCode(HttpStatus.SC_CONFLICT);
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void canProjectsBeFetchedViaService() throws EntityExistsException {
        Project projectOfUser1 = new Project();
        projectOfUser1.setUserId("user1");
        projectService.add(projectOfUser1);
        Project projectOfUser2 = new Project();
        projectOfUser2.setUserId("user2");
        projectService.add(projectOfUser2);

        given()
                .contentType("application/json")
                .get("/api/v1/projects")
                .then()
                .statusCode(200)
                .body(
                        "projects.size()", is(1),
                        "projects.id", hasItems(projectOfUser1.getId().toString()));
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void canSpecificProjectBeFetchedViaService() throws EntityExistsException {
        Project project1 = new Project();
        project1.setUserId("user1");
        projectService.add(project1);
        Project project2 = new Project();
        project2.setUserId("user1");
        projectService.add(project2);

        given()
                .contentType("application/json")
                .get(String.format("/api/v1/projects/%s", project2.getId()))
                .then()
                .statusCode(200)
                .body(
                        "id", equalTo(project2.getId().toString()));
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void fetchingAProjectFailsIfItDoesNotExist() throws EntityExistsException {
        Project project = new Project();
        project.setUserId("user1");

        given()
                .contentType("application/json")
                .get(String.format("/api/v1/projects/%s", project.getId()))
                .then()
                .statusCode(404);
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void fetchingAProjectFailsIfItDoesNotBelongToUser() throws EntityExistsException {
        Project project = new Project();
        project.setUserId("user2");
        projectService.add(project);

        given()
                .contentType("application/json")
                .get(String.format("/api/v1/projects/%s", project.getId()))
                .then()
                .statusCode(404);
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void projectCanBeDeletedViaService() throws EntityExistsException {
        Project project = new Project();
        project.setUserId("user1");
        projectService.add(project);

        Assertions.assertEquals(1, projectService.listAll().size());

        given()
                .contentType("application/json")
                .delete(String.format("/api/v1/projects/%s", project.getId()))
                .then()
                .statusCode(200);

        Assertions.assertEquals(0, projectService.listAll().size());
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void deletingAProjectFailsIfItDoesNotExist() throws EntityExistsException {
        Project project = new Project();
        project.setUserId("user1");

        given()
                .contentType("application/json")
                .delete(String.format("/api/v1/projects/%s", project.getId()))
                .then()
                .statusCode(404);
    }

    @Test
    @TestSecurity(user = "user1", roles = { "user" })
    public void deletingAProjectFailsIfItDoesNotBelongToUser() throws EntityExistsException {
        Project project = new Project();
        project.setUserId("user2");
        projectService.add(project);

        Assertions.assertEquals(1, projectService.listAll().size());

        given()
                .contentType("application/json")
                .delete(String.format("/api/v1/projects/%s", project.getId()))
                .then()
                .statusCode(404);

        Assertions.assertEquals(1, projectService.listAll().size());
    }
}
