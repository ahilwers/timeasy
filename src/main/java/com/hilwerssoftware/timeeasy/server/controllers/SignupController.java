package com.hilwerssoftware.timeeasy.server.controllers;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/signup")
public class SignupController {

    @Autowired
    private AccountRepository accountRepository;

    @PostMapping
    public String signup(@RequestBody Account account) {
        accountRepository.insert(account);
        return "OK";
    }

    @GetMapping("/test")
    public String getTest() {
        return "Test";
    }
}
