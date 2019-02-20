package irt.systemx.crypto;

public class KeyPairing {

    public String privPairing;
    public String g1Pub;
    public String g2Pub;

    public KeyPairing() {
	this.setPrivPairing("-1");
	this.setG1Pub("-1");
	this.setG2Pub("-1");
    }

    public String getPrivPairing() {
	return privPairing;
    }

    public void setPrivPairing(String privPairing) {
	this.privPairing = privPairing;
    }

    public String getG1Pub() {
	return g1Pub;
    }

    public void setG1Pub(String g1Pub) {
	this.g1Pub = g1Pub;
    }

    public String getG2Pub() {
	return g2Pub;
    }

    public void setG2Pub(String g2Pub) {
	this.g2Pub = g2Pub;
    }

}
