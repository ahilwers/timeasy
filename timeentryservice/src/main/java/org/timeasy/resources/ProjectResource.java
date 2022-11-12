package org.timeasy.resources;

import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.Consumes;
import javax.ws.rs.POST;
import javax.ws.rs.Path;

import org.eclipse.microprofile.jwt.JsonWebToken;
import org.timeasy.models.Project;
import org.timeasy.services.ProjectService;
import org.timeasy.services.UserDataService;
import org.timeasy.tools.EntityExistsException;

import io.quarkus.security.Authenticated;

@Path("/api/v1/projects")
@Authenticated
public class ProjectResource {

    private final ProjectService projectService;
    private final JsonWebToken token;
    private final UserDataService userDataService;

    public ProjectResource(ProjectService projectService, JsonWebToken token, UserDataService userDataService) {
        this.projectService = projectService;
        this.token = token;
        this.userDataService = userDataService;
    }

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    @Produces(MediaType.APPLICATION_JSON)
    public EntityCreationInfo addProject(Project project) throws EntityExistsException {
        String userId = userDataService.getUserId(token);
        project.setUserId(userId);
        projectService.add(project);
        return new EntityCreationInfo(project.getId().toString());
    }

}
