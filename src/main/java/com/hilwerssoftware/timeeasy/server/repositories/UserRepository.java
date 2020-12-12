package com.hilwerssoftware.timeeasy.server.repositories;

import com.hilwerssoftware.timeeasy.server.models.User;
import org.springframework.data.mongodb.repository.MongoRepository;

public interface UserRepository  extends MongoRepository<User, String> {
}
