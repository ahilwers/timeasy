package com.hilwerssoftware.timeeasy.server.exceptions;

public class OwnerNotInDatabaseException extends Exception {
    public OwnerNotInDatabaseException() {
        super("Owner does not exist in database. Please create it first.");
    }

    public OwnerNotInDatabaseException(String message) {
        super(message);
    }
}
