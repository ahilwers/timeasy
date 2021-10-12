package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.exceptions.OwnerMissingException;
import com.hilwerssoftware.timeeasy.server.exceptions.OwnerNotInDatabaseException;
import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public interface TimeEntryService {

    void addTimeEntry(TimeEntry timeEntry) throws OwnerMissingException, OwnerNotInDatabaseException;
    List<TimeEntry> getTimeEntries(String ownerId);

}
