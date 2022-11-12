package org.timeasy.services;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

import javax.inject.Inject;
import javax.transaction.Transactional;

import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.tools.EntityExistsException;
import org.timeasy.tools.EntityNotFoundException;

import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;

@QuarkusTest
@Testcontainers
public class ProjectServiceTest {

    @InjectMock
    SecurityIdentity securityIdentity;
    @Inject
    private ProjectService projectService;
    @Inject
    private ProjectRepository projectRepository;

    @BeforeEach
    @Transactional
    public void setup() {
        projectRepository.deleteAll();
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
    }

    @Test
    public void canProjectBeAdded() throws EntityExistsException {
        Project project = new Project();
        project.setDescription("Testproject");
        projectService.add(project);

        Project projectFromDb = projectRepository.findById(project.getId());
        Assertions.assertEquals(project.getId().toString(), projectFromDb.getId().toString());
        Assertions.assertEquals(project.getDescription(), projectFromDb.getDescription());
    }

    @Test
    public void addingAProjectFailsIfItAlreadyExists() throws EntityExistsException {
        Project firstProject = new Project();
        firstProject.setDescription("First Project");
        projectService.add(firstProject);

        Project secondProject = new Project();
        secondProject.setId(firstProject.getId());
        firstProject.setDescription("Second Project");

        Assertions.assertThrows(EntityExistsException.class, () -> {
            projectService.add(secondProject);
        });
    }

    @Test
    public void canProjectBeUpdated() throws EntityExistsException, EntityNotFoundException {
        Project project = new Project();
        project.setDescription("Project");
        project.setUserId("1");
        project.setDeleted(false);
        projectService.add(project);

        // Need to re-fetch because the timestamp in the database is not as precise as
        // the Instant.
        Instant createdTimeStamp = projectService.findById(project.getId()).getCreatedTimeStamp();

        Project projectToBeUpdated = new Project();
        projectToBeUpdated.setId(project.getId());
        projectToBeUpdated.setDescription("Updated description");
        projectToBeUpdated.setUserId("2");
        projectToBeUpdated.setDeleted(project.isDeleted());
        projectToBeUpdated.setCreatedTimeStamp(project.getCreatedTimeStamp());
        projectToBeUpdated.setUpdatedTimeStamp(project.getUpdatedTimeStamp());

        projectService.update(projectToBeUpdated);

        List<Project> projectList = projectRepository.listAll();
        Assertions.assertEquals(1, projectList.size());
        Project updatedProject = projectList.get(0);
        Assertions.assertEquals(projectToBeUpdated.getDescription(), updatedProject.getDescription());
        Assertions.assertEquals(projectToBeUpdated.getUserId(), updatedProject.getUserId());
        Assertions.assertEquals(createdTimeStamp, updatedProject.getCreatedTimeStamp());
        Assertions.assertNotEquals(project.getUpdatedTimeStamp(), updatedProject.getUpdatedTimeStamp());
    }

    @Test
    public void updatingAProjectFailsIfItDoesNotExist() {
        Project project = new Project();
        Assertions.assertThrows(EntityNotFoundException.class, () -> {
            projectService.update(project);
        });
    }

    @Test
    @Transactional
    public void canListOfAllProjectBeFetched() throws EntityExistsException {
        createProjects(10);
        List<Project> projectList = projectService.listAll();
        Assertions.assertEquals(10, projectList.size());
        for (int i = 0; i < 10; i++) {
            Project project = projectList.get(i);
            Assertions.assertEquals(String.format("Project %s", i), project.getDescription());
        }
    }

    @Test
    public void deleteProjectAreNotContainedInProjectList() throws EntityNotFoundException, EntityExistsException {
        createProjects(2);
        createProjects(1, 2, true);
        List<Project> projectsFromDb = projectService.listAll();
        Assertions.assertEquals(2, projectsFromDb.size());
        for (Project project : projectsFromDb) {
            Assertions.assertNotEquals("Project 2", project.getDescription());
        }
    }

    @Test
    public void canSpecificProjectBeFetched() throws EntityExistsException {
        List<Project> projectList = createProjects(3);
        Project projectFromDb = projectService.findById(projectList.get(1).getId());
        Assertions.assertEquals(projectList.get(1).getId(), projectFromDb.getId());
        Assertions.assertEquals(projectList.get(1).getDescription(), projectFromDb.getDescription());
    }

    @Test
    public void canProjectsOfSpecificUserBeFetched() throws EntityExistsException {
        createProjects(3);
        List<Project> projectsOfUser = projectService.listAllOfUser("user 1");
        Assertions.assertEquals(1, projectsOfUser.size());
        Assertions.assertEquals("Project 1", projectsOfUser.get(0).getDescription());
    }

    private List<Project> createProjects(int count) throws EntityExistsException {
        return createProjects(count, 0, false);
    }

    private List<Project> createProjects(int count, int startIndex, boolean deleted) throws EntityExistsException {
        List<Project> projectList = new ArrayList<>();
        for (int i = 0; i < count; i++) {
            Project project = new Project();
            project.setDescription(String.format("Project %s", startIndex + i));
            project.setUserId(String.format("user %s", startIndex + i));
            project.setDeleted(deleted);
            projectService.add(project);
            projectList.add(project);
        }
        return projectList;
    }
}
