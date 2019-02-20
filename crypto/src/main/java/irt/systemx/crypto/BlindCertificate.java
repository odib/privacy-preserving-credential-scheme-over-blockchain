package irt.systemx.crypto;

public class BlindCertificate {

    String blindCommitment; 
    String blindCertificate; 
    String blindPubG1CP; 
    String blindPubG2User;
    String blindPrivUser;
    String blindGenerator;
    String blindFactor;
    
    public BlindCertificate() {
	this.setBlindCommitment("-1"); 
	this.setBlindCertificate("-1"); 
	this.setBlindPubG1CP("-1"); 
	this.setBlindPubG2User("-1");
	this.setBlindPrivUser("-1");
	this.setBlindGenerator("-1");
	this.setBlindFactor("-1");
    }
    
    public String getBlindCommitment() {
        return blindCommitment;
    }
    public void setBlindCommitment(String blindCommitment) {
        this.blindCommitment = blindCommitment;
    }
    public String getBlindCertificate() {
        return blindCertificate;
    }
    public void setBlindCertificate(String blindCertificate) {
        this.blindCertificate = blindCertificate;
    }
    public String getBlindPubG1CP() {
        return blindPubG1CP;
    }
    public void setBlindPubG1CP(String blindPubG1CP) {
        this.blindPubG1CP = blindPubG1CP;
    }
    public String getBlindPubG2User() {
        return blindPubG2User;
    }
    public void setBlindPubG2User(String blindPubG2User) {
        this.blindPubG2User = blindPubG2User;
    }
    public String getBlindPrivUser() {
        return blindPrivUser;
    }
    public void setBlindPrivUser(String blindPrivUser) {
        this.blindPrivUser = blindPrivUser;
    }
    public String getBlindGenerator() {
        return blindGenerator;
    }
    public void setBlindGenerator(String blindGenerator) {
        this.blindGenerator = blindGenerator;
    }
    public String getBlindFactor() {
        return blindFactor;
    }
    public void setBlindFactor(String blindFactor) {
        this.blindFactor = blindFactor;
    }
    
    
    
}
