package com.hilwerssoftware.timeeasy.server.repositories;

import com.hilwerssoftware.timeeasy.server.models.Account;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.data.mongodb.repository.Query;

import java.util.Optional;

public interface AccountRepository extends MongoRepository<Account, String> {

    @Query("{ 'username': ?0 }")
    Optional<Account> findByUsername(String username);

}
