package com.hilwerssoftware.timeasy.models;

import org.hibernate.annotations.Type;

import javax.persistence.Entity;
import javax.persistence.Id;
import java.time.Instant;
import java.util.UUID;

@Entity
public class TimeEntry {

    @Id
    @Type(type = "pg-uuid")
    public UUID id;

    public String name;

    public Instant startTime;

    public TimeEntry() {
        id = UUID.randomUUID();
    }
}
