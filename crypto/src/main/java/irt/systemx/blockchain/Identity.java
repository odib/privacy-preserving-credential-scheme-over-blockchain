package irt.systemx.blockchain;

public class Identity {
    private String certificate;

    public String getCertificate() {
	return certificate;
    }

    public void setCertificate(String certificate) {
	this.certificate = certificate;
    }

    @Override
    public String toString() {
	return "ClassPojo [certificate = " + certificate + "]";
    }
}
