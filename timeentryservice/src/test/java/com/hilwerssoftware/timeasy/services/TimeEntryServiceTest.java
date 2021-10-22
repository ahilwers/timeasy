package com.hilwerssoftware.timeasy.services;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import com.hilwerssoftware.timeasy.repositories.TimeEntryRepository;
import com.hilwerssoftware.timeasy.tools.EntityExistsException;
import com.hilwerssoftware.timeasy.tools.EntityNotFoundException;
import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.testcontainers.junit.jupiter.Testcontainers;

import javax.inject.Inject;
import javax.transaction.Transactional;
import java.time.Instant;
import java.util.List;

@QuarkusTest
@Testcontainers
public class TimeEntryServiceTest {

    @InjectMock
    SecurityIdentity securityIdentity;
    @Inject
    TimeEntryService timeEntryService;
    @Inject
    TimeEntryRepository timeEntryRepository;

    @BeforeEach
    @Transactional
    public void setup() {
        timeEntryRepository.deleteAll();
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
    }

    @Test
    public void canTimeEntrybeAdded() throws EntityExistsException {
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.setDescription("Testentry");
        timeEntryService.add(timeEntry);

        TimeEntry timeEntryFromDb = timeEntryService.findById(timeEntry.getId());
        Assertions.assertEquals(timeEntry.getId().toString(), timeEntryFromDb.getId().toString());
        Assertions.assertEquals("Testentry", timeEntryFromDb.getDescription());
    }

    @Test
    public void addingTimeEntryFailsIfTimeEntryAlreadyExists() throws EntityExistsException {
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.setDescription("ExistingEntry");
        timeEntryService.add(timeEntry);

        TimeEntry newTimeEntry = new TimeEntry();
        newTimeEntry.setId(timeEntry.getId());
        newTimeEntry.setDescription("SomeName");
        Assertions.assertThrows(EntityExistsException.class, () -> {
            timeEntryService.add(newTimeEntry);
        });
    }

    @Test
    public void canTimeEntryBeUpdated() throws EntityExistsException, EntityNotFoundException {
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.setDescription("TimeEntry1");
        timeEntry.setStartTime(Instant.parse("1980-04-09T10:15:30.00Z"));
        timeEntry.setEndTime(Instant.parse("1980-04-09T10:16:30.00Z"));
        timeEntry.setProjectId("project1");
        timeEntry.setUserId("1");
        timeEntryService.add(timeEntry);
        // As the timestamp in the database is not as precise as the instant, we must re-fetch the time entry and store
        // the created timestamp for later reference.
        Instant createdTimeStamp = timeEntryService.findById(timeEntry.getId()).getCreatedTimeStamp();

        TimeEntry timeEntryToBeUpdated = new TimeEntry();
        timeEntryToBeUpdated.setId(timeEntry.getId());
        timeEntryToBeUpdated.setDescription("UpdatedTimeEntry");
        Instant newStartTime = Instant.parse("1975-12-28T12:17:40.00Z");
        timeEntryToBeUpdated.setStartTime(newStartTime);
        Instant newEndTime = Instant.parse("1975-12-28T14:17:40.00Z");
        timeEntryToBeUpdated.setEndTime(newEndTime);
        timeEntryToBeUpdated.setProjectId("project2");
        timeEntryToBeUpdated.setUserId("2");
        timeEntryToBeUpdated.setCreatedTimeStamp(timeEntry.getCreatedTimeStamp());
        timeEntryToBeUpdated.setUpdatedTimeStamp(timeEntry.getUpdatedTimeStamp());
        timeEntryService.update(timeEntryToBeUpdated);

        List<TimeEntry> timeEntries = timeEntryRepository.listAll();
        Assertions.assertEquals(1, timeEntries.size());
        TimeEntry updatedTimeEntry = timeEntries.get(0);
        Assertions.assertEquals(timeEntry.getId(), updatedTimeEntry.getId());
        Assertions.assertEquals("UpdatedTimeEntry", updatedTimeEntry.getDescription());
        Assertions.assertEquals(newStartTime.toString(), updatedTimeEntry.getStartTime().toString());
        Assertions.assertEquals(newEndTime.toString(), updatedTimeEntry.getEndTime().toString());
        Assertions.assertEquals("project2", updatedTimeEntry.getProjectId());
        Assertions.assertEquals("2", updatedTimeEntry.getUserId());
        Assertions.assertEquals(createdTimeStamp, updatedTimeEntry.getCreatedTimeStamp(), "Created time stamps do not match.");
        Assertions.assertNotEquals(timeEntry.getUpdatedTimeStamp(), updatedTimeEntry.getUpdatedTimeStamp(), "Updated timestamps must not match!");
    }

    @Test
    public void updatingATimeEntryFailsIfItDoesNotExist() {
        TimeEntry timeEntry = new TimeEntry();
        Assertions.assertThrows(EntityNotFoundException.class, () -> {
            timeEntryService.update(timeEntry);
        });
    }


}
