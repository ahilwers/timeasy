package com.hilwerssoftware.timeeasy.server.configuration;

import com.hilwerssoftware.timeeasy.server.repositories.AccountRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.authentication.builders.AuthenticationManagerBuilder;
import org.springframework.security.config.annotation.authentication.configuration.GlobalAuthenticationConfigurerAdapter;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;

@Configuration
public class AuthorizationConfiguration extends GlobalAuthenticationConfigurerAdapter{

    @Autowired
    AccountRepository accountRepository;

    @Override
    public void init(AuthenticationManagerBuilder auth) throws Exception {
        auth.userDetailsService(getUserDetailsService());
    }

    @Bean
    UserDetailsService getUserDetailsService() {
        return new UserDetailsService() {
            @Override
            public UserDetails loadUserByUsername(String username) throws UsernameNotFoundException {
                var foundUser = accountRepository.findByUsername(username);
                if (!foundUser.isPresent()) {
                    throw new UsernameNotFoundException(String.format("A user with the name %s was not found.", username));
                }
                return foundUser.get();
            }
        };
    }
}
