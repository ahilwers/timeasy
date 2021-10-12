package com.hilwerssoftware.timeeasy.server.services;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class AccountServiceImpl implements AccountService {

    @Autowired
    private AccountRepository accountRepository;

    @Autowired
    private PasswordEncoder passwordEncoder;

    @Override
    public void addAccount(Account account) {
        account.setPassword(encryptPassword(account.getPassword()));
        accountRepository.insert(account);
    }

    @Override
    public boolean accountExists(String userName) {
        return findByUsername(userName).isPresent();
    }

    @Override
    public Optional<Account> findByUsername(String userName) {
        return accountRepository.findByUsername(userName);
    }

    private String encryptPassword(String password) {
        return passwordEncoder.encode(password);
    }
}
