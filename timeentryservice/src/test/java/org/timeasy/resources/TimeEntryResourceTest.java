package org.timeasy.resources;

import org.timeasy.models.TimeEntry;
import org.timeasy.repositories.TimeEntryRepository;
import org.timeasy.services.TimeEntryService;
import org.timeasy.services.UserDataService;
import org.timeasy.tools.EntityExistsException;
import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;
import io.vertx.core.json.JsonObject;
import org.apache.http.HttpStatus;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import javax.inject.Inject;
import javax.transaction.Transactional;
import javax.ws.rs.core.MediaType;

import java.util.List;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.hasItems;
import static org.hamcrest.CoreMatchers.is;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;

@QuarkusTest
public class TimeEntryResourceTest {

    @InjectMock
    SecurityIdentity securityIdentity;
    @InjectMock
    UserDataService userDataService;
    @Inject
    TimeEntryRepository timeEntryRepository;
    @Inject
    TimeEntryService timeEntryService;

    @BeforeEach
    @Transactional
    public void setup() {
        timeEntryRepository.deleteAll();
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
        Mockito.when(userDataService.getUserId(any())).thenReturn("user1");
    }

    @Test
    public void canTimeEntryBeAddedViaService() {
        JsonObject jsonObject = new JsonObject()
                .put("description", "Timeentry")
                .put("projectid", "project1");
        given()
                .contentType(MediaType.APPLICATION_JSON)
                .body(jsonObject.toString())
                .when()
                .post("/api/v1/timeentries")
                .then()
                .assertThat()
                .statusCode(HttpStatus.SC_OK);
        List<TimeEntry> timeEntries = timeEntryService.listAll();
        assertEquals(1, timeEntries.size());
        TimeEntry timeEntry = timeEntries.get(0);
        assertEquals("Timeentry", timeEntry.getDescription());
        // the time entry should belong to the correct user:
        assertEquals("user1", timeEntry.getUserId());
    }

    @Test
    public void addingATimeEntryViaServiceFailsIfTimeEntryExists() throws EntityExistsException {
        TimeEntry timeEntry = new TimeEntry();
        timeEntryService.add(timeEntry);
        JsonObject jsonObject = new JsonObject()
                .put("id", timeEntry.getId());
        given()
                .contentType(MediaType.APPLICATION_JSON)
                .body(jsonObject.toString())
                .when()
                .post("/api/v1/timeentries")
                .then()
                .assertThat()
                .statusCode(HttpStatus.SC_CONFLICT);
    }

    @Test
    public void canTimEntriesBeFetchedViaService() throws EntityExistsException {
        TimeEntry entryOfUser1 = new TimeEntry();
        entryOfUser1.setUserId("user1");
        timeEntryService.add(entryOfUser1);
        TimeEntry entryOfUser2 = new TimeEntry();
        entryOfUser2.setUserId("user2");
        timeEntryService.add(entryOfUser2);

        given()
                .contentType("application/json")
                .get("/api/v1/timeentries")
                .then()
                .statusCode(200)
                .body(
                    "timeEntries.size()", is(1),
                    "timeEntries.id", hasItems(entryOfUser1.getId().toString())
                );
    }

    @Test
    public void canTimEntriesOfProjectBeFetchedViaService() throws EntityExistsException {
        TimeEntry entryOfUser1 = new TimeEntry();
        entryOfUser1.setUserId("user1");
        entryOfUser1.setProjectId("project1");
        timeEntryService.add(entryOfUser1);
        TimeEntry entryOfUser2 = new TimeEntry();
        entryOfUser2.setUserId("user2");
        entryOfUser2.setProjectId("project1");
        timeEntryService.add(entryOfUser2);
        TimeEntry secondEntryOfUser1 = new TimeEntry();
        secondEntryOfUser1.setUserId("user1");
        secondEntryOfUser1.setProjectId("project2");
        timeEntryService.add(secondEntryOfUser1);

        given()
                .contentType("application/json")
                .queryParam("project", "project1")
                .get("/api/v1/timeentries")
                .then()
                .statusCode(200)
                .body(
                        "timeEntries.size()", is(1),
                        "timeEntries.id", hasItems(entryOfUser1.getId().toString())
                );
    }
}
