package org.timeasy.repositories;

import java.util.UUID;

import javax.enterprise.context.ApplicationScoped;

import org.timeasy.models.TimeEntry;

import io.quarkus.hibernate.orm.panache.PanacheRepositoryBase;

@ApplicationScoped
public class TimeEntryRepository implements PanacheRepositoryBase<TimeEntry, UUID> {
}
