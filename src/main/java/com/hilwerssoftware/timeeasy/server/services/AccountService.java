package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.models.Account;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public interface AccountService {

    void addAccount(Account account);

    boolean accountExists(String userName);

    Optional<Account> findByUsername(String userName);
}
