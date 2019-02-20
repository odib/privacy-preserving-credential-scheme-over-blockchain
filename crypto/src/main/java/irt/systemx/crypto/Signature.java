package irt.systemx.crypto;

public class Signature {

    public String sign; // s
    public String rand; // r

    public Signature() {
	this.setSign("-1");
	this.setRand("-1");
    }

    public String getSign() {
	return sign;
    }

    public void setSign(String sign) {
	this.sign = sign;
    }

    public String getRand() {
	return rand;
    }

    public void setRand(String rand) {
	this.rand = rand;
    }

}