package com.hilwerssoftware.timeeasy.server.repositories;

import com.hilwerssoftware.timeeasy.server.models.Account;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.data.mongodb.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface AccountRepository extends MongoRepository<Account, String> {

    @Query("{ 'username': ?0 }")
    Optional<Account> findByUsername(String username);

}
