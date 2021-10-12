package com.hilwerssoftware.timeeasy.server.timeentries;

import com.hilwerssoftware.timeeasy.server.exceptions.OwnerMissingException;
import com.hilwerssoftware.timeeasy.server.exceptions.OwnerNotInDatabaseException;
import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import com.hilwerssoftware.timeeasy.server.repositories.TimeEntryRepository;
import com.hilwerssoftware.timeeasy.server.services.TimeEntryService;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.util.Assert;

@SpringBootTest
public class TimeEntryServiceTest {

    @Autowired
    private TimeEntryService timeEntryService;

    @Autowired
    private TimeEntryRepository timeEntryRepository;

    @Autowired
    private AccountRepository accountRepository;

    @Test
    public void canTimeEntryBeAdded() throws OwnerMissingException, OwnerNotInDatabaseException {
        Account account = new Account();
        accountRepository.insert(account);
        TimeEntry timeEntry = new TimeEntry();
        timeEntry.setOwner(account);
        timeEntryService.addTimeEntry(timeEntry);
        Assert.hasText(timeEntry.getId(), "The time entry needs to have an id.");
        var addedTimeEntry = timeEntryRepository.findById(timeEntry.getId());
        Assert.isTrue(addedTimeEntry.isPresent(), "addedTimeEntry was not found.");
    }

    @Test
    public void onlyTimeEntriesOfCurrentUserAreReturned() {
        var account1 = new Account();
        accountRepository.insert(account1);
        var account2 = new Account();
        accountRepository.insert(account2);

        var timeEntryOfAccount1 = new TimeEntry();
        timeEntryOfAccount1.setOwner(account1);
        timeEntryRepository.insert(timeEntryOfAccount1);

        var timeEntryOfAccount2 = new TimeEntry();
        timeEntryOfAccount2.setOwner(account2);
        timeEntryRepository.insert(timeEntryOfAccount2);

        var timeEntries = timeEntryService.getTimeEntries(account1.getId());
        Assert.notEmpty(timeEntries, "The time entries of account 1 should not be emppty.");
        Assert.state(timeEntries.size()==1, "The amount of time entries for account 1 should be 1");
        Assert.state(timeEntries.get(0).getOwner().getId().equals(account1.getId()), "The time entry should belong to account 1.");

        timeEntries = timeEntryService.getTimeEntries(account2.getId());
        Assert.notEmpty(timeEntries, "The time entries of account 2 should not be emppty.");
        Assert.state(timeEntries.size()==1, "The amount of time entries for account 2 should be 1");
        Assert.state(timeEntries.get(0).getOwner().getId().equals(account2.getId()), "The time entry should belong to account 2.");
    }

    @Test
    public void addingTimeEntryWithoutOwnerFails() {
        Assertions.assertThrows(OwnerMissingException.class, () -> {
            var timeEntry = new TimeEntry();
            timeEntryService.addTimeEntry(timeEntry);
        });
    }

    @Test
    public void addingTimeEntryWithNonExistingOwnerFails() {
        Assertions.assertThrows(OwnerNotInDatabaseException.class, () -> {
           var timeEntry = new TimeEntry();
           timeEntry.setOwner(new Account());
           timeEntryService.addTimeEntry(timeEntry);
        });
    }
}
