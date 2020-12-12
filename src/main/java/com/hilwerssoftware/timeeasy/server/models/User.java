package com.hilwerssoftware.timeeasy.server.models;

import org.springframework.data.annotation.Id;

public class User {

    public enum UserType {
        PRIVATE,
        ORGANIZATION
    }

    @Id
    public String id;
    private String username;
    private String password;
    private UserType userType;


    public String getId() {
        return id;
    }

    protected void setId(String id) {
        this.id = id;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public UserType getUserType() {
        return userType;
    }

    public void setUserType(UserType userType) {
        this.userType = userType;
    }
}
