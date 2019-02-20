package irt.systemx.crypto;

public class Key {

    public String privKey;
    public String pubKey;

    public Key() {
	this.setPrivKey("-1");
	this.setPubKey("-1");
    }

    public Key(String privKey, String pubKey) {
	this.setPrivKey(privKey);
	this.setPubKey(pubKey);
    }

    public String getPubKey() {
	return pubKey;
    }

    public void setPubKey(String pubKey) {
	this.pubKey = pubKey;
    }

    public String getPrivKey() {
	return privKey;
    }

    public void setPrivKey(String privKey) {
	this.privKey = privKey;
    }

}
