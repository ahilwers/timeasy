package org.timeasy.resources;

import java.util.List;

import org.timeasy.models.Project;

/**
 * Wrapper class to be returned by the service.
 * We use this class to avoid returning the array directly because this way
 * we're able to add more informations later
 * on if we need to.
 */
public class ProjectList {

    private List<Project> projects;

    public ProjectList(List<Project> projects) {
        this.projects = projects;
    }

    public List<Project> getProjects() {
        return projects;
    }

    protected void setProjects(List<Project> projects) {
        this.projects = projects;
    }
}
