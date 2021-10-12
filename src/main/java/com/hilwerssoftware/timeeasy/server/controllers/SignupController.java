package com.hilwerssoftware.timeeasy.server.controllers;

import com.hilwerssoftware.timeeasy.server.models.Account;
import com.hilwerssoftware.timeeasy.server.services.AccountService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/signup")
public class SignupController {

    @Autowired
    private AccountService accountService;

    @PostMapping
    public String signup(@RequestBody Account account) {
        accountService.addAccount(account);
        return "OK";
    }

    @GetMapping("/test")
    public String getTest() {
        return "Test";
    }
}
