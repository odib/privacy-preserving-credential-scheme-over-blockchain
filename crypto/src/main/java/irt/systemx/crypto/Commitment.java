package irt.systemx.crypto;

public class Commitment {

    public String commit; 
    public String rand;
    
    public Commitment() {
	this.setCommit("-1");
	this.setRand("-1");
    }
    
    public Commitment(String commit, String rand) {
	this.setCommit(commit);
	this.setRand(rand);
    }
    public String getCommit() {
        return commit;
    }
    public void setCommit(String commit) {
        this.commit = commit;
    }
    public String getRand() {
        return rand;
    }
    public void setRand(String rand) {
        this.rand = rand;
    }
    
    
}
