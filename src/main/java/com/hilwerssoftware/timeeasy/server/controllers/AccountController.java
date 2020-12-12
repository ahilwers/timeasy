package com.hilwerssoftware.timeeasy.server.controllers;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class AccountController {

    @Autowired
    private AccountRepository accountRepository;

    @RequestMapping("/test")
    public String test() {
        return "Greetings from Timeasy!";
    }

    @RequestMapping("/admin/test")
    public String adminTest() {
        return "Hello admin!";
    }

}
