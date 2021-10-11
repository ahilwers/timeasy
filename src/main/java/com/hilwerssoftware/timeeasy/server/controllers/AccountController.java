package com.hilwerssoftware.timeeasy.server.controllers;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/v1/accounts")
public class AccountController {

    @Autowired
    private AccountRepository accountRepository;

    @GetMapping("/test")
    public String test() {
        return "Greetings from Timeasy!";
    }

    @GetMapping("/admin/test")
    public String adminTest() {
        return "Hello admin!";
    }

    @GetMapping("")
    public List<Account> getAllAccounts() {
        return accountRepository.findAll();
    }


}
