package irt.systemx.crypto;

public class ZKP {

    public String A;
    public String t;
    public String pubSecret;

    public ZKP() {
	this.setA("-1");
	this.setT("-1");
	this.setPubSecret("-1");
    }

    public String getA() {
	return A;
    }

    public void setA(String a) {
	A = a;
    }

    public String getT() {
	return t;
    }

    public void setT(String t) {
	this.t = t;
    }

    public String getPubSecret() {
	return pubSecret;
    }

    public void setPubSecret(String pubSecret) {
	this.pubSecret = pubSecret;
    }

}
