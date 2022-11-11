package org.timeasy.services;

import javax.enterprise.context.ApplicationScoped;
import javax.transaction.Transactional;

import org.timeasy.models.Project;
import org.timeasy.repositories.ProjectRepository;
import org.timeasy.tools.EntityExistsException;

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

}
