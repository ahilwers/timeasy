package com.hilwerssoftware.timeeasy.server.initialization;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.services.AccountService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.security.crypto.password.PasswordEncoder;

public class ProductionDataInitializer implements ApplicationListener<ContextRefreshedEvent> {

    @Autowired
    private AccountService accountService;

    @Override
    public void onApplicationEvent(ContextRefreshedEvent event) {
        initializeRootAccount();
    }

    private void initializeRootAccount() {

        if (!accountService.accountExists("root")) {
            Account rootAccount = new Account();
            rootAccount.setUsername("root");
            rootAccount.setPassword("root");
            rootAccount.setUserRole(Account.UserRole.ADMIN);
            accountService.addAccount(rootAccount);
        }
    }
}
