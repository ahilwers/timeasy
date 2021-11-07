package org.timeasy.resources;

import org.timeasy.models.TimeEntry;

import java.util.List;

/**
 * Wrapper class to be returned by the service.
 * We use this class to avoid returning the array directly because this way we're abled to add more informations later
 * on if we need to.
 */
public class TimeEntryList {

    private List<TimeEntry> timeEntries;

    public TimeEntryList(List<TimeEntry> timeEntries) {
        this.timeEntries = timeEntries;
    }

    public List<TimeEntry> getTimeEntries() {
        return timeEntries;
    }

    protected void setTimeEntries(List<TimeEntry> timeEntries) {
        this.timeEntries = timeEntries;
    }
}
