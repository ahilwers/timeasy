package com.hilwerssoftware.timeasy;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import com.hilwerssoftware.timeasy.repositories.TimeEntryRepository;
import org.jboss.resteasy.annotations.jaxrs.PathParam;

import javax.inject.Inject;
import javax.transaction.Transactional;
import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import java.time.Instant;
import java.util.UUID;

@Path("/hello")
public class GreetingResource {

    @Inject
    TimeEntryRepository timeEntryRepository;

    @GET
    @Produces(MediaType.TEXT_PLAIN)
    @Transactional
    public String hello() {
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.name = "Testname";
        timeEntry.startTime = Instant.now();
        timeEntryRepository.persist(timeEntry);
        return String.format("Timeentry created - Id: %s", timeEntry.id);
    }


    @GET
    @Path("/update/{id}")
    @Produces(MediaType.TEXT_PLAIN)
    @Transactional
    public String update(@PathParam String id) {
        TimeEntry timeEntry = timeEntryRepository.findById(UUID.fromString(id));
        timeEntry.name = "Testname2";
        return String.format("Enitty %s updated", id);
    }
}