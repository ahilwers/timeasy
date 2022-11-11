package org.timeasy.models;

import java.time.Instant;
import java.util.UUID;

import javax.persistence.Entity;
import javax.persistence.Id;

import org.hibernate.annotations.Type;

@Entity
public class Project {
  @Id
  @Type(type = "pg-uuid")
  private UUID id;
  private String description;
  private String userId;
  private Instant createdTimeStamp = Instant.now();
  private Instant updatedTimeStamp = Instant.now();
  private boolean deleted;

  public Project() {
    id = UUID.randomUUID();
  }

  public UUID getId() {
    return id;
  }

  public void setId(UUID id) {
    this.id = id;
  }

  public String getDescription() {
    return description;
  }

  public void setDescription(String description) {
    this.description = description;
  }

  public String getUserId() {
    return userId;
  }

  public void setUserId(String userId) {
    this.userId = userId;
  }

  public Instant getCreatedTimeStamp() {
    return createdTimeStamp;
  }

  public void setCreatedTimeStamp(Instant createdTimeStamp) {
    this.createdTimeStamp = createdTimeStamp;
  }

  public Instant getUpdatedTimeStamp() {
    return updatedTimeStamp;
  }

  public void setUpdatedTimeStamp(Instant updatedTimeStamp) {
    this.updatedTimeStamp = updatedTimeStamp;
  }

  public boolean isDeleted() {
    return deleted;
  }

  public void setDeleted(boolean deleted) {
    this.deleted = deleted;
  }
}
