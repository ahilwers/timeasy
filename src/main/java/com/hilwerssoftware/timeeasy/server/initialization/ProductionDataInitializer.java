package com.hilwerssoftware.timeeasy.server.initialization;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.security.crypto.password.PasswordEncoder;

public class ProductionDataInitializer implements ApplicationListener<ContextRefreshedEvent> {

    @Autowired
    private AccountRepository accountRepository;
    @Autowired
    private PasswordEncoder passwordEncoder;

    @Override
    public void onApplicationEvent(ContextRefreshedEvent event) {
        initializeRootAccount();
    }

    private void initializeRootAccount() {
        var rootUser = accountRepository.findByUsername("root");
        if (rootUser.isEmpty()) {
            Account rootAccount = new Account();
            rootAccount.setUsername("root");
            rootAccount.setPassword(passwordEncoder.encode("root"));
            rootAccount.setUserRole(Account.UserRole.ADMIN);
            accountRepository.insert(rootAccount);
        }
    }
}
