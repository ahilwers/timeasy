package com.hilwerssoftware.timeasy.services;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import io.quarkus.security.identity.SecurityIdentity;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.junit.mockito.InjectMock;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import javax.inject.Inject;

@QuarkusTest
public class TimeEntryServiceTest {

    @InjectMock
    SecurityIdentity securityIdentity;
    @Inject
    TimeEntryService timeEntryService;

    @BeforeEach
    public void setup() {
        Mockito.when(securityIdentity.hasRole("user")).thenReturn(true);
    }

    @Test
    public void canTimeEntrybeAdded() {
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.name = "Testentry";
        timeEntryService.add(timeEntry);

        TimeEntry timeEntryFromDb = timeEntryService.findById(timeEntry.id);
        Assertions.assertEquals(timeEntry.id.toString(), timeEntryFromDb.id.toString());
        Assertions.assertEquals("Testentry", timeEntryFromDb.name);
    }


}
