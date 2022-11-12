package org.timeasy.services;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

import javax.enterprise.context.ApplicationScoped;
import javax.persistence.LockModeType;
import javax.transaction.Transactional;

import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.tools.EntityExistsException;
import org.timeasy.tools.EntityNotFoundException;

import io.quarkus.panache.common.Sort;

@ApplicationScoped
public class ProjectService {

    private ProjectRepository projectRepository;

    public ProjectService(ProjectRepository projectRepository) {
        this.projectRepository = projectRepository;
    }

    @Transactional
    public void add(Project project) throws EntityExistsException {
        if (projectRepository.findByIdOptional(project.getId()).isPresent()) {
            throw new EntityExistsException(String.format("A project with the id %s already exists.", project.getId()));
        }
        projectRepository.persist(project);
    }

    @Transactional
    public void update(Project project) throws EntityNotFoundException {
        doUpdate(project);
    }

    private void doUpdate(Project project) throws EntityNotFoundException {
        Project existingProject = projectRepository.findById(project.getId(), LockModeType.PESSIMISTIC_WRITE);
        if (existingProject == null) {
            throw new EntityNotFoundException(
                    String.format("A Project with the id %s could not be found.", project.getId()));
        }
        existingProject.setId(project.getId());
        existingProject.setDescription(project.getDescription());
        existingProject.setUserId(project.getUserId());
        existingProject.setDeleted(project.isDeleted());
        existingProject.setCreatedTimeStamp(project.getCreatedTimeStamp());
        existingProject.setUpdatedTimeStamp(Instant.now());
    }

    @Transactional
    public Project findById(UUID id) {
        return projectRepository.findById(id);
    }

    @Transactional
    public List<Project> listAll() {
        return projectRepository.list("deleted", Sort.by("description"), false);
    }

}
