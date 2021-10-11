package com.hilwerssoftware.timeeasy.server.repositories;

import com.hilwerssoftware.timeeasy.server.models.TimeEntry;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.data.mongodb.repository.Query;

import java.util.List;

public interface TimeEntryRepository extends MongoRepository<TimeEntry, String> {

    @Query("{ 'owner.id': ?0 }")
    List<TimeEntry> findForOwner(String ownerId);

}
