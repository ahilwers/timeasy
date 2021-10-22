package com.hilwerssoftware.timeasy.resources;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import com.hilwerssoftware.timeasy.services.TimeEntryService;
import io.quarkus.security.Authenticated;
import org.jboss.resteasy.annotations.jaxrs.PathParam;

import javax.inject.Inject;
import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import java.util.UUID;

@Path("/api/v1/timeentries")
@Authenticated
public class TimeEntryResource {

    @Inject
    TimeEntryService timeEntryService;

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    @Path("/{id}")
    public TimeEntry getTimeEntry(@PathParam String id) {
        return timeEntryService.findById(UUID.fromString(id));
    }

}
