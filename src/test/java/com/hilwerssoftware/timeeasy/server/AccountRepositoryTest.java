package com.hilwerssoftware.timeeasy.server;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.util.Assert;

import static org.assertj.core.api.Assertions.*;

@SpringBootTest
public class AccountRepositoryTest {

    @Autowired
    AccountRepository accountRepository;

    @Test
    public void canAccountBeAdded() {
        Account account = new Account();
        account.setUsername("test");
        account.setPassword("test");
        accountRepository.insert(account);

        // There should be two accounts now as the root account ist created during startup:
        var accounts = accountRepository.findAll();
        assertThat(accounts.size()).isEqualTo(2);
        // Try to find the account we just created:
        var accountFromDb = accountRepository.findById(account.getId());
        assertThat(accountFromDb.isEmpty()).isFalse();
        assertThat(accountFromDb.get().getUsername()).isEqualTo(account.getUsername());
        assertThat(accountFromDb.get().getPassword()).isEqualTo(account.getPassword());
    }
}
