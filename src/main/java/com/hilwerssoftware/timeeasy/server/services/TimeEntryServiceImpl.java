package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.exceptions.OwnerMissingException;
import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import com.hilwerssoftware.timeeasy.server.repositories.TimeEntryRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class TimeEntryServiceImpl implements TimeEntryService {

    @Autowired
    private TimeEntryRepository timeEntryRepository;

    public void addTimeEntry(TimeEntry timeEntry) throws OwnerMissingException {
        if (timeEntry.getOwner()==null) {
            throw new OwnerMissingException();
        }
        timeEntryRepository.insert(timeEntry);
    }

    public List<TimeEntry> getTimeEntries(String ownerId) {
        return timeEntryRepository.findForOwner(ownerId);
    }

}
