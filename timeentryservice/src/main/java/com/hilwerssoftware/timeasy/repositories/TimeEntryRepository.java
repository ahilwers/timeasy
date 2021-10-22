package com.hilwerssoftware.timeasy.repositories;

import com.hilwerssoftware.timeasy.models.TimeEntry;
import io.quarkus.hibernate.orm.panache.PanacheRepositoryBase;

import javax.enterprise.context.ApplicationScoped;
import java.util.UUID;

@ApplicationScoped
public class TimeEntryRepository implements PanacheRepositoryBase<TimeEntry, UUID> {
}
