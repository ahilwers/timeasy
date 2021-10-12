package com.hilwerssoftware.timeeasy.server.exceptions;

public class OwnerMissingException extends Exception {

    public OwnerMissingException() {
        super("The entity has now owner.");
    }

    public OwnerMissingException(String message) {
        super(message);
    }
}
