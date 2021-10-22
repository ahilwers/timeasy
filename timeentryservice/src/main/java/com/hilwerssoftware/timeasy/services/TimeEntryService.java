package com.hilwerssoftware.timeasy.services;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import com.hilwerssoftware.timeasy.repositories.TimeEntryRepository;

import javax.enterprise.context.ApplicationScoped;
import javax.transaction.Transactional;
import java.util.UUID;

@ApplicationScoped
public class TimeEntryService {

    private TimeEntryRepository timeEntryRepository;

    public TimeEntryService(TimeEntryRepository timeEntryRepository) {
        this.timeEntryRepository = timeEntryRepository;
    }

    @Transactional
    public void add(TimeEntry timeEntry) {
        timeEntryRepository.persist(timeEntry);
    }

    @Transactional
    public TimeEntry findById(UUID id) {
        return timeEntryRepository.findById(id);
    }

}
