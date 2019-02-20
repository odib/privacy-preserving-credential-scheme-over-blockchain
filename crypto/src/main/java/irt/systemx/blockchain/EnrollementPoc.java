package irt.systemx.blockchain;

import java.io.Serializable;
import java.security.KeyPair;
import java.security.PrivateKey;

import org.hyperledger.fabric.sdk.Enrollment;

public class EnrollementPoc implements Enrollment, Serializable {
    private static final long serialVersionUID = 550416591376968096L;
    private KeyPair key;
    private String cert;

    public EnrollementPoc(KeyPair signingKeyPair, String signedPem) {
	key = signingKeyPair;
	this.cert = signedPem;
    }

    @Override
    public PrivateKey getKey() {
	return key.getPrivate();
    }

    @Override
    public String getCert() {
	return cert;
    }

}