package irt.systemx.blockchain;

public class EnrollmentUser {
    private Identity identity;
    private String signingIdentity;

    public Identity getIdentity() {
	return identity;
    }

    public void setIdentity(Identity identity) {
	this.identity = identity;
    }

    public String getSigningIdentity() {
	return signingIdentity;
    }

    public void setSigningIdentity(String signingIdentity) {
	this.signingIdentity = signingIdentity;
    }

    @Override
    public String toString() {
	return "ClassPojo [identity = " + identity + ", signingIdentity = " + signingIdentity + "]";
    }
}
