package org.timeasy.resources;

import org.timeasy.models.TimeEntry;
import org.timeasy.services.TimeEntryService;
import org.timeasy.services.UserDataService;
import org.timeasy.tools.EntityExistsException;
import io.quarkus.security.Authenticated;
import io.quarkus.security.identity.SecurityIdentity;
import org.eclipse.microprofile.jwt.JsonWebToken;
import org.jboss.resteasy.annotations.jaxrs.PathParam;
import org.jboss.resteasy.annotations.jaxrs.QueryParam;

import javax.ws.rs.*;
import javax.ws.rs.core.MediaType;
import java.util.UUID;

@Path("/api/v1/timeentries")
@Authenticated
public class TimeEntryResource {

    private final UserDataService userDataService;
    private JsonWebToken token;
    private TimeEntryService timeEntryService;
    private SecurityIdentity securityIdentity;

    public TimeEntryResource(TimeEntryService timeEntryService, SecurityIdentity securityIdentity, JsonWebToken token, UserDataService userDataService) {
        this.timeEntryService = timeEntryService;
        this.securityIdentity = securityIdentity;
        this.token = token;
        this.userDataService = userDataService;
    }

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    public TimeEntryList getTimeEntries(@QueryParam String project) {
        String userId = userDataService.getUserId(token);
        if (project!=null)
            return new TimeEntryList(timeEntryService.listAllOfUserAndProject(userId, project));
        else
            return new TimeEntryList(timeEntryService.listAllOfUser(userId));
    }

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    @Path("/{id}")
    public TimeEntry getTimeEntry(@PathParam String id) {
        return timeEntryService.findById(UUID.fromString(id));
    }

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    @Produces(MediaType.APPLICATION_JSON)
    public TimeEntryCreationInfo addTimeEntry(TimeEntry timeEntry) throws EntityExistsException {
        String userId = userDataService.getUserId(token);
        timeEntry.setUserId(userId);
        timeEntryService.add(timeEntry);
        return new TimeEntryCreationInfo(timeEntry.getId().toString());
    }

}
