package org.timeasy.services;

import org.eclipse.microprofile.jwt.JsonWebToken;

import javax.enterprise.context.ApplicationScoped;

@ApplicationScoped
public class UserDataService {

    public String getUserId(JsonWebToken token) {
        return token.getClaim("sub");
    }
}
