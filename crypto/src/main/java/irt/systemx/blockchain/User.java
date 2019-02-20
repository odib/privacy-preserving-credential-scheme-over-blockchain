package irt.systemx.blockchain;

public class User {
    private String roles;
    private String name;
    private EnrollmentUser enrollment;
    private String affiliation;
    private String mspid;
    private String algo;
    private String enrollmentSecret;

    public String getRoles() {
	return roles;
    }

    public void setRoles(String roles) {
	this.roles = roles;
    }

    public String getName() {
	return name;
    }

    public void setName(String name) {
	this.name = name;
    }

    public EnrollmentUser getEnrollment() {
	return enrollment;
    }

    public void setEnrollment(EnrollmentUser enrollment) {
	this.enrollment = enrollment;
    }

    public String getAffiliation() {
	return affiliation;
    }

    public void setAffiliation(String affiliation) {
	this.affiliation = affiliation;
    }

    public String getMspid() {
	return mspid;
    }

    public void setMspid(String mspid) {
	this.mspid = mspid;
    }

    public String getEnrollmentSecret() {
	return enrollmentSecret;
    }

    public void setEnrollmentSecret(String enrollmentSecret) {
	this.enrollmentSecret = enrollmentSecret;
    }

    public String getAlgo() {
	return algo;
    }

    public void setAlgo(String algo) {
	this.algo = algo;
    }
    
    @Override
    public String toString() {
	return "ClassPojo [roles = " + roles + ", name = " + name + ", enrollment = " + enrollment + ", affiliation = "
		+ affiliation + ", mspid = " + mspid + ", enrollmentSecret = " + enrollmentSecret + "]";
    }
}