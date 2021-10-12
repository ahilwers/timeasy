package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.exceptions.OwnerMissingException;
import com.hilwerssoftware.timeeasy.server.exceptions.OwnerNotInDatabaseException;
import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import com.hilwerssoftware.timeeasy.server.repositories.TimeEntryRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class TimeEntryServiceImpl implements TimeEntryService {

    @Autowired
    private TimeEntryRepository timeEntryRepository;

    @Autowired
    private AccountService accountService;

    public void addTimeEntry(TimeEntry timeEntry) throws OwnerMissingException, OwnerNotInDatabaseException {
        if (timeEntry.getOwner()==null) {
            throw new OwnerMissingException();
        }
        if (!accountService.accountExists(timeEntry.getOwner().getUsername())) {
            throw new OwnerNotInDatabaseException();
        }
        timeEntryRepository.insert(timeEntry);
    }

    public List<TimeEntry> getTimeEntries(String ownerId) {
        return timeEntryRepository.findForOwner(ownerId);
    }

}
