package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import com.hilwerssoftware.timeeasy.server.repositories.TimeEntryRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class TimeEntryService {

    @Autowired
    private TimeEntryRepository timeEntryRepository;

    public void addTimeEntry(TimeEntry timeEntry) {
        timeEntryRepository.insert(timeEntry);
    }

    public List<TimeEntry> getTimeEntries(String ownerId) {
        return timeEntryRepository.findForOwner(ownerId);
    }

}
