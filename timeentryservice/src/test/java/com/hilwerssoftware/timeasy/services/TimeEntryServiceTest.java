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
import java.sql.Time;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Comparator;
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

    @Test
    @Transactional
    public void canListOfAllTimeEntriesBeFetched() {
        createTimeEntries(10);
        List<TimeEntry> timeEntries = timeEntryService.listAll();
        Assertions.assertEquals(10, timeEntries.size());
        timeEntries.sort(new Comparator<TimeEntry>() {
            @Override
            public int compare(TimeEntry timeEntry1, TimeEntry timeEntry2) {
                return timeEntry1.getDescription().compareTo(timeEntry2.getDescription());
            }
        });
        for (int i=0; i<10; i++) {
            TimeEntry timeEntry = timeEntries.get(i);
            Assertions.assertEquals(String.format("Timeentry %s", i), timeEntry.getDescription());
        }
    }

    @Test
    @Transactional
    public void canSpecificTimeEntryBeFetched() {
        List<TimeEntry> timeEntries = createTimeEntries(3);
        TimeEntry timeEntryFromDb = timeEntryService.findById(timeEntries.get(1).getId());
        Assertions.assertEquals(timeEntries.get(1).getId(), timeEntryFromDb.getId());
        Assertions.assertEquals("Timeentry 1", timeEntryFromDb.getDescription());
    }


    private List<TimeEntry> createTimeEntries(int count) {
        List<TimeEntry> timeEntries = new ArrayList<>();
        for (int i=0; i<count; i++) {
            TimeEntry timeEntry = new TimeEntry();
            timeEntry.setDescription(String.format("Timeentry %s", i));
            timeEntry.setUserId(String.format("user %s", i));
            timeEntry.setProjectId(String.format("project %s", i));
            timeEntryRepository.persist(timeEntry);
            timeEntries.add(timeEntry);
        }
        return timeEntries;
    }

    @Test
    @Transactional
    public void canTimeEntriesOfSpecificUserBeFetched() {
        List<TimeEntry> timeEntries = createTimeEntries(3);
        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfUser("user 1");
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(timeEntries.get(1).getId(), timeEntriesOfUser.get(0).getId());
        Assertions.assertEquals("user 1", timeEntriesOfUser.get(0).getUserId());
    }

    @Test
    @Transactional
    public void canTimeEntriesOfSpecificProjectBeFetched() {
        List<TimeEntry> timeEntries = createTimeEntries(3);
        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfProject("project 1");
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(timeEntries.get(1).getId(), timeEntriesOfUser.get(0).getId());
        Assertions.assertEquals("project 1", timeEntriesOfUser.get(0).getProjectId());
    }

    @Test
    public void canTimeEntriesOfUserAndProjectBeFetched() throws EntityExistsException {
        TimeEntry timeEntry1 = new TimeEntry();
        timeEntry1.setUserId("user1");
        timeEntry1.setProjectId("project1");
        timeEntryService.add(timeEntry1);
        TimeEntry timeEntry2 = new TimeEntry();
        timeEntry2.setUserId("user1");
        timeEntry2.setProjectId("project2");
        timeEntryService.add(timeEntry2);
        TimeEntry timeEntry3 = new TimeEntry();
        timeEntry3.setUserId("user2");
        timeEntry3.setProjectId("project2");
        timeEntryService.add(timeEntry3);
        List<TimeEntry> timeEntries = timeEntryService.listAllOfUserAndProject("user1", "project2");
        Assertions.assertEquals(1, timeEntries.size());
        Assertions.assertEquals(timeEntry2.getId(), timeEntries.get(0).getId());
        Assertions.assertEquals("user1", timeEntries.get(0).getUserId());
        Assertions.assertEquals("project2", timeEntries.get(0).getProjectId());
    }

    @Test
    public void canTimeEntrybeDeleted() throws EntityExistsException, EntityNotFoundException {
        TimeEntry timeEntry1 = new TimeEntry();
        timeEntry1.setUserId("user1");
        timeEntry1.setProjectId("project1");
        timeEntryService.add(timeEntry1);
        TimeEntry timeEntry2 = new TimeEntry();
        timeEntry2.setUserId("user1");
        timeEntry2.setProjectId("project2");
        timeEntryService.add(timeEntry2);
        TimeEntry timeEntry3 = new TimeEntry();
        timeEntry3.setUserId("user2");
        timeEntry3.setProjectId("project2");
        timeEntryService.add(timeEntry3);

        Assertions.assertEquals(3, timeEntryService.listAll().size());

        timeEntryService.delete(timeEntry2);
        List<TimeEntry> allTimeEntries = timeEntryService.listAll();
        Assertions.assertEquals(2, allTimeEntries.size());
        for (TimeEntry timeEntry: allTimeEntries) {
            Assertions.assertNotEquals(timeEntry2.getId(), timeEntry.getId());
        }

        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfUser("user1");
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(timeEntry1.getId(), timeEntriesOfUser.get(0).getId());

        List<TimeEntry> timeEntriesOfProject = timeEntryService.listAllOfProject("project2");
        Assertions.assertEquals(1, timeEntriesOfProject.size());
        Assertions.assertEquals(timeEntry3.getId(), timeEntriesOfProject.get(0).getId());

        List<TimeEntry> timeEntriesOfUserAndProject = timeEntryService.listAllOfUserAndProject("user1", "project2");
        Assertions.assertEquals(0, timeEntriesOfUserAndProject.size());
    }

    @Test
    public void deletingATimeEntryFailsIfItDoesNotExist() {
        TimeEntry timeEntry = new TimeEntry();
        Assertions.assertThrows(EntityNotFoundException.class, () -> {
           timeEntryService.delete(timeEntry);
        });
    }


}
