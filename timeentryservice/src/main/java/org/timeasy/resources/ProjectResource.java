package org.timeasy.resources;

import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

import java.util.UUID;

import javax.ws.rs.Consumes;
import javax.ws.rs.DELETE;
import javax.ws.rs.GET;
import javax.ws.rs.POST;
import javax.ws.rs.Path;

import org.eclipse.microprofile.jwt.JsonWebToken;
import org.timeasy.models.Project;
import org.timeasy.services.ProjectService;
import org.timeasy.services.UserDataService;
import org.timeasy.tools.EntityExistsException;
import org.timeasy.tools.EntityNotFoundException;

import io.quarkus.security.Authenticated;
import org.jboss.resteasy.annotations.jaxrs.PathParam;

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

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    public ProjectList getProjects() {
        String userId = userDataService.getUserId(token);
        return new ProjectList(projectService.listAllOfUser(userId));
    }

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    @Path("/{id}")
    public Project getProject(@PathParam String id) throws EntityNotFoundException {
        Project project = projectService.findById(UUID.fromString(id));
        String userId = userDataService.getUserId(token);
        if (!project.getUserId().equals(userId)) {
            throw new EntityNotFoundException(String.format("A project with the id %s could not be found.", id));
        }
        return project;
    }

    @DELETE
    @Produces(MediaType.APPLICATION_JSON)
    @Path("/{id}")
    public Response deleteProject(@PathParam String id) throws EntityNotFoundException {
        Project project = projectService.findById(UUID.fromString(id));
        String userId = userDataService.getUserId(token);
        if (!project.getUserId().equals(userId)) {
            throw new EntityNotFoundException(String.format("A project with the id %s could not be found.", id));
        }
        projectService.delete(project);
        return Response.status(Response.Status.OK).build();
    }

}
