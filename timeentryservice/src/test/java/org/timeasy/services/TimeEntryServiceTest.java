package org.timeasy.services;

import org.timeasy.models.Project;
import org.timeasy.models.TimeEntry;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.repositories.TimeEntryRepository;
import org.timeasy.tools.EntityExistsException;
import org.timeasy.tools.EntityNotFoundException;
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
    @Inject
    ProjectService projectService;
    @Inject
    ProjectRepository projectRepository;

    @BeforeEach
    @Transactional
    public void setup() {
        timeEntryRepository.deleteAll();
        projectRepository.deleteAll();
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
        Project project1 = new Project();
        project1.setDescription("Project 1");
        projectService.add(project1);
        Project project2 = new Project();
        project2.setDescription("Project 2");
        projectService.add(project2);

        TimeEntry timeEntry = new TimeEntry();
        timeEntry.setDescription("TimeEntry1");
        timeEntry.setStartTime(Instant.parse("1980-04-09T10:15:30.00Z"));
        timeEntry.setEndTime(Instant.parse("1980-04-09T10:16:30.00Z"));
        timeEntry.setProject(project1);
        timeEntry.setUserId("1");
        timeEntryService.add(timeEntry);
        // As the timestamp in the database is not as precise as the instant, we must
        // re-fetch the time entry and store
        // the created timestamp for later reference.
        Instant createdTimeStamp = timeEntryService.findById(timeEntry.getId()).getCreatedTimeStamp();

        TimeEntry timeEntryToBeUpdated = new TimeEntry();
        timeEntryToBeUpdated.setId(timeEntry.getId());
        timeEntryToBeUpdated.setDescription("UpdatedTimeEntry");
        Instant newStartTime = Instant.parse("1975-12-28T12:17:40.00Z");
        timeEntryToBeUpdated.setStartTime(newStartTime);
        Instant newEndTime = Instant.parse("1975-12-28T14:17:40.00Z");
        timeEntryToBeUpdated.setEndTime(newEndTime);
        timeEntryToBeUpdated.setProject(project2);
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
        Assertions.assertEquals("Project 2", updatedTimeEntry.getProject().getDescription());
        Assertions.assertEquals("2", updatedTimeEntry.getUserId());
        Assertions.assertEquals(createdTimeStamp, updatedTimeEntry.getCreatedTimeStamp(),
                "Created time stamps do not match.");
        Assertions.assertNotEquals(timeEntry.getUpdatedTimeStamp(), updatedTimeEntry.getUpdatedTimeStamp(),
                "Updated timestamps must not match!");
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
    public void canListOfAllTimeEntriesBeFetched() throws EntityExistsException {
        createTimeEntries(10);
        List<TimeEntry> timeEntries = timeEntryService.listAll();
        Assertions.assertEquals(10, timeEntries.size());
        timeEntries.sort(new Comparator<TimeEntry>() {
            @Override
            public int compare(TimeEntry timeEntry1, TimeEntry timeEntry2) {
                return timeEntry1.getDescription().compareTo(timeEntry2.getDescription());
            }
        });
        for (int i = 0; i < 10; i++) {
            TimeEntry timeEntry = timeEntries.get(i);
            Assertions.assertEquals(String.format("Timeentry %s", i), timeEntry.getDescription());
        }
    }

    @Test
    @Transactional
    public void canSpecificTimeEntryBeFetched() throws EntityExistsException {
        TestData testData = createTimeEntries(3);
        TimeEntry timeEntryFromDb = timeEntryService.findById(testData.getTimeEntries().get(1).getId());
        Assertions.assertEquals(testData.getTimeEntries().get(1).getId(), timeEntryFromDb.getId());
        Assertions.assertEquals("Timeentry 1", timeEntryFromDb.getDescription());
    }

    private TestData createTimeEntries(int count) throws EntityExistsException {
        List<TimeEntry> timeEntries = new ArrayList<>();
        List<Project> projects = new ArrayList<>();
        for (int i = 0; i < count; i++) {
            Project project = new Project();
            project.setDescription(String.format("Project %s", i));
            projects.add(project);
            projectService.add(project);

            TimeEntry timeEntry = new TimeEntry();
            timeEntry.setDescription(String.format("Timeentry %s", i));
            timeEntry.setUserId(String.format("user %s", i));
            timeEntry.setProject(project);
            timeEntryRepository.persist(timeEntry);
            timeEntries.add(timeEntry);
        }
        return new TestData(projects, timeEntries);
    }

    @Test
    @Transactional
    public void canTimeEntriesOfSpecificUserBeFetched() throws EntityExistsException {
        TestData testData = createTimeEntries(3);
        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfUser("user 1");
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(testData.getTimeEntries().get(1).getId(), timeEntriesOfUser.get(0).getId());
        Assertions.assertEquals("user 1", timeEntriesOfUser.get(0).getUserId());
    }

    @Test
    @Transactional
    public void canTimeEntriesOfSpecificProjectBeFetched() throws EntityExistsException {
        TestData testData = createTimeEntries(3);
        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfProject(testData.getProjects().get(1).getId());
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(testData.getTimeEntries().get(1).getId(), timeEntriesOfUser.get(0).getId());
        Assertions.assertEquals("Project 1", timeEntriesOfUser.get(0).getProject().getDescription());
    }

    @Test
    public void canTimeEntriesOfUserAndProjectBeFetched() throws EntityExistsException {
        Project project1 = new Project();
        project1.setDescription("Project 1");
        projectService.add(project1);

        Project project2 = new Project();
        project2.setDescription("Project 2");
        projectService.add(project2);

        TimeEntry timeEntry1 = new TimeEntry();
        timeEntry1.setUserId("user1");
        timeEntry1.setProject(project1);
        timeEntryService.add(timeEntry1);
        TimeEntry timeEntry2 = new TimeEntry();
        timeEntry2.setUserId("user1");
        timeEntry2.setProject(project2);
        timeEntryService.add(timeEntry2);
        TimeEntry timeEntry3 = new TimeEntry();
        timeEntry3.setUserId("user2");
        timeEntry3.setProject(project2);
        timeEntryService.add(timeEntry3);
        List<TimeEntry> timeEntries = timeEntryService.listAllOfUserAndProject("user1", project2.getId());
        Assertions.assertEquals(1, timeEntries.size());
        Assertions.assertEquals(timeEntry2.getId(), timeEntries.get(0).getId());
        Assertions.assertEquals("user1", timeEntries.get(0).getUserId());
        Assertions.assertEquals("Project 2", timeEntries.get(0).getProject().getDescription());
    }

    @Test
    public void canTimeEntrybeDeleted() throws EntityExistsException, EntityNotFoundException {
        Project project1 = new Project();
        project1.setDescription("Project 1");
        projectService.add(project1);

        Project project2 = new Project();
        project2.setDescription("Project 2");
        projectService.add(project2);

        TimeEntry timeEntry1 = new TimeEntry();
        timeEntry1.setUserId("user1");
        timeEntry1.setProject(project1);
        timeEntryService.add(timeEntry1);
        TimeEntry timeEntry2 = new TimeEntry();
        timeEntry2.setUserId("user1");
        timeEntry2.setProject(project2);
        timeEntryService.add(timeEntry2);
        TimeEntry timeEntry3 = new TimeEntry();
        timeEntry3.setUserId("user2");
        timeEntry3.setProject(project2);
        timeEntryService.add(timeEntry3);

        Assertions.assertEquals(3, timeEntryService.listAll().size());

        timeEntryService.delete(timeEntry2);
        List<TimeEntry> allTimeEntries = timeEntryService.listAll();
        Assertions.assertEquals(2, allTimeEntries.size());
        for (TimeEntry timeEntry : allTimeEntries) {
            Assertions.assertNotEquals(timeEntry2.getId(), timeEntry.getId());
        }

        List<TimeEntry> timeEntriesOfUser = timeEntryService.listAllOfUser("user1");
        Assertions.assertEquals(1, timeEntriesOfUser.size());
        Assertions.assertEquals(timeEntry1.getId(), timeEntriesOfUser.get(0).getId());

        List<TimeEntry> timeEntriesOfProject = timeEntryService.listAllOfProject(project2.getId());
        Assertions.assertEquals(1, timeEntriesOfProject.size());
        Assertions.assertEquals(timeEntry3.getId(), timeEntriesOfProject.get(0).getId());

        List<TimeEntry> timeEntriesOfUserAndProject = timeEntryService.listAllOfUserAndProject("user1",
                project2.getId());
        Assertions.assertEquals(0, timeEntriesOfUserAndProject.size());
    }

    @Test
    public void deletingATimeEntryFailsIfItDoesNotExist() {
        TimeEntry timeEntry = new TimeEntry();
        Assertions.assertThrows(EntityNotFoundException.class, () -> {
            timeEntryService.delete(timeEntry);
        });
    }

    class TestData {

        private List<Project> projects;
        private List<TimeEntry> timeEntries;

        public TestData(List<Project> projects, List<TimeEntry> timeEntries) {
            this.projects = projects;
            this.timeEntries = timeEntries;
        }

        public List<Project> getProjects() {
            return projects;
        }

        public List<TimeEntry> getTimeEntries() {
            return timeEntries;
        }
    }
}
