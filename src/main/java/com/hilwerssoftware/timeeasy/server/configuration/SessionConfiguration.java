package com.hilwerssoftware.timeeasy.server.configuration;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.session.data.mongo.config.annotation.web.http.EnableMongoHttpSession;
import org.springframework.session.web.context.AbstractHttpSessionApplicationInitializer;
import org.springframework.session.web.http.HeaderHttpSessionIdResolver;
import org.springframework.session.web.http.HttpSessionIdResolver;

@Configuration
@EnableMongoHttpSession
public class SessionConfiguration extends AbstractHttpSessionApplicationInitializer {

    @Bean
    public HttpSessionIdResolver httpSessionIdResolver() {
        // Return the session as header value "X-Auth-Token" instead of setting a cookie. Needed for REST authentication
        return HeaderHttpSessionIdResolver.xAuthToken();
    }

}
