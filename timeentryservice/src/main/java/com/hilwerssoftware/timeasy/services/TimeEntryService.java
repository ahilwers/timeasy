package com.hilwerssoftware.timeasy.services;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import com.hilwerssoftware.timeasy.repositories.TimeEntryRepository;
import com.hilwerssoftware.timeasy.tools.EntityExistsException;
import com.hilwerssoftware.timeasy.tools.EntityNotFoundException;

import javax.enterprise.context.ApplicationScoped;
import javax.persistence.LockModeType;
import javax.transaction.Transactional;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

@ApplicationScoped
public class TimeEntryService {

    private TimeEntryRepository timeEntryRepository;

    public TimeEntryService(TimeEntryRepository timeEntryRepository) {
        this.timeEntryRepository = timeEntryRepository;
    }

    @Transactional
    public void add(TimeEntry timeEntry) throws EntityExistsException {
        if (timeEntryRepository.findByIdOptional(timeEntry.getId()).isPresent()) {
            throw new EntityExistsException(String.format("A time entry with the id %s already exists.", timeEntry.getId()));
        }
        timeEntryRepository.persist(timeEntry);
    }

    @Transactional
    public void update(TimeEntry timeEntry) throws EntityNotFoundException {
        doUpdate(timeEntry);
    }

    private void doUpdate(TimeEntry timeEntry) throws EntityNotFoundException {
        TimeEntry existingTimeentry = timeEntryRepository.findById(timeEntry.getId(), LockModeType.PESSIMISTIC_WRITE);
        if (existingTimeentry==null) {
            throw new EntityNotFoundException(String.format("A time entry with the id %s does not exist.", timeEntry.getId().toString()));
        }
        existingTimeentry.setId(timeEntry.getId());
        existingTimeentry.setDescription(timeEntry.getDescription());
        existingTimeentry.setStartTime(timeEntry.getStartTime());
        existingTimeentry.setEndTime(timeEntry.getEndTime());
        existingTimeentry.setProjectId(timeEntry.getProjectId());
        existingTimeentry.setUserId(timeEntry.getUserId());
        existingTimeentry.setDeleted(timeEntry.isDeleted());
        existingTimeentry.setCreatedTimeStamp(timeEntry.getCreatedTimeStamp());
        existingTimeentry.setUpdatedTimeStamp(Instant.now());
    }

    @Transactional
    public TimeEntry findById(UUID id) {
        return timeEntryRepository.findById(id);
    }


    @Transactional
    public List<TimeEntry> listAll() {
        return timeEntryRepository.list("deleted", false);
    }

    @Transactional
    public List<TimeEntry> listAllOfUser(String userId) {
        return timeEntryRepository.list("userid=?1 and deleted=?2", userId, false);
    }

    @Transactional
    public List<TimeEntry> listAllOfProject(String projectId) {
        return timeEntryRepository.list("projectid=?1 and deleted=?2", projectId, false);
    }

    @Transactional
    public List<TimeEntry> listAllOfUserAndProject(String userId, String projectId) {
        return timeEntryRepository.list("userid=?1 and projectid=?2 and deleted=?3", userId, projectId, false);
    }

    @Transactional
    public void delete(TimeEntry timeEntry) throws EntityNotFoundException {
        timeEntry.setDeleted(true);
        doUpdate(timeEntry);
    }
}
