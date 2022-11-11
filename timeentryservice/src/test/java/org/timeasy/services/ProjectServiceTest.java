package org.timeasy.services;

import javax.transaction.Transactional;

import javax.inject.Inject;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.mockito.Mockito;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.tools.EntityExistsException;

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
}
