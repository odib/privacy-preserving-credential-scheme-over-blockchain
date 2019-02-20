package irt.systemx.main;

import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.util.Map;

import org.apache.http.entity.StringEntity;
import org.apache.log4j.Logger;
import org.hyperledger.fabric.sdk.Channel;
import org.hyperledger.fabric.sdk.HFClient;
import org.json.simple.JSONObject;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import irt.systemx.blockchain.HTTPClient;
import irt.systemx.crypto.BlindCertificate;
import irt.systemx.crypto.Commitment;
import irt.systemx.crypto.Key;
import irt.systemx.crypto.KeyPairing;
import irt.systemx.crypto.Signature;
import irt.systemx.crypto.ZKP;
import irt.systemx.blockchain.BlockhainCommunicator;

public class MainProtocol {

    private static ObjectMapper mapper = new ObjectMapper();
    private static final Logger LOG = Logger.getLogger(MainProtocol.class);
    private static final String SERVER_IP = "127.0.0.1";
    private static final String SERVER_PORT = "8000";
    private static final String MAIN_URL = "http://" + SERVER_IP + ":" + SERVER_PORT;
    private static final String GET_KEYS_API_URL = MAIN_URL + "/user/generateKey";
    private static final String COMMITMENT_API_URL = MAIN_URL + "/user/commitment";
    private static final String IV_SIGN_API_URL = MAIN_URL + "/iv/signCommitment";
    private static final String IV_VERIFY_API_URL = MAIN_URL + "/iv/verifySignature";
    private static final String GENERATE_ZKP_RANDOM_API_URL = MAIN_URL + "/user/generateZKP/random";
    private static final String GENERATE_ZKP_AGE_API_URL = MAIN_URL + "/user/generateZKP/age";
    private static final String VERIFY_PROOF_RANDOM_API_URL = MAIN_URL + "/CP/verifyProof/random";
    private static final String VERIFY_PROOF_AGE_API_URL = MAIN_URL + "/CP/verifyProof/age";
    private static final String GENERATE_KEY_PAIRING_API_URL = MAIN_URL + "/user/generateKeyPairing";
    private static final String GENERATE_CERTIFICATE_API_URL = MAIN_URL + "/CP/generateCertificate";
    private static final String VERIFY_CERTIFICATE_API_URL = MAIN_URL + "/user/verifyCertificate";
    private static final String BLIND_CERTIFICATE_API_URL = MAIN_URL + "/user/blindCertificate";
    private static final String VERIFY_BLIND_CERTIFICATE_API_URL = MAIN_URL + "/SP/verifyBlindCertificate";
    private static final String CHANNEL_NAME = "mychannel";
    private static final String CHAINCODE_NAME = "aav";//anonymous-attribute-verifier";
    private static final String CHAINCODE_FUNCTION = "verify";//Verify before granting access 

    public static void main(String[] args) throws UnsupportedEncodingException, Exception {

	// Define the attribute to be blinded/protected
	String age = "21";

	// Generating Keys
	// #1 The user generates a key pairs to commit his age attribute
	Key userKeys = getKeys(GET_KEYS_API_URL);
	// #2 Generating signature keys for the Identity Validator
	Key ivKeys = getKeys(GET_KEYS_API_URL);
	// #3 Generating pairing key pairs for the User and Certificate Provider
	KeyPairing kpUser = generateKeyPairing(GENERATE_KEY_PAIRING_API_URL);
	KeyPairing kpCP = generateKeyPairing(GENERATE_KEY_PAIRING_API_URL);

	// Compute the commitment for the age attribute
	Commitment com = commit(userKeys.getPubKey(), age, COMMITMENT_API_URL);

	// The Identity Validator signs the User's commitment
	Signature sign = ivSign(com.getCommit(), ivKeys.getPrivKey(), ivKeys.getPubKey(), IV_SIGN_API_URL);

	// The user verifies that the signature is authentic
	ivVerify(sign.getSign(), sign.getRand(), ivKeys.getPubKey(), com.getCommit(), IV_VERIFY_API_URL);

	// Generating a ZKP for the random number used in the commitment phase
	ZKP zkRand = generateZKP(com.getRand(), userKeys.getPubKey(), GENERATE_ZKP_RANDOM_API_URL);

	// Generating a ZKP for the age attribute used in the commitment phase
	ZKP zkAge = generateZKP(age, userKeys.getPubKey(), GENERATE_ZKP_AGE_API_URL);

	// The Certificate Provider checks the validity of the above proofs
	verifyProof(zkRand.getA(), zkRand.getT(), userKeys.getPubKey(), zkRand.getPubSecret(),
		VERIFY_PROOF_RANDOM_API_URL);
	verifyProof(zkAge.getA(), zkAge.getT(), userKeys.getPubKey(), zkAge.getPubSecret(), VERIFY_PROOF_AGE_API_URL);

	// The certificate Provider generates a certificate for the User using the
	// Blinded attributes of the user
	String certificate = generateCertificate(com.getCommit(), kpCP.getPrivPairing(), kpUser.getG2Pub(),
		GENERATE_CERTIFICATE_API_URL);

	// The User checks the validity of the certificate
	verifyCertificate(com.getCommit(), certificate, kpCP.getG1Pub(), kpUser.getG2Pub(), VERIFY_CERTIFICATE_API_URL);

	// The User self blinds his certificate
	BlindCertificate bc = blindCertificate(com.getCommit(), certificate, kpCP.getG1Pub(), kpUser.getG2Pub(),
		kpUser.getPrivPairing(), BLIND_CERTIFICATE_API_URL);

	// The Service Provider checks the validity of the User's certificate
	verifyBlindCertificate(bc.getBlindCommitment(), bc.getBlindPubG1CP(), bc.getBlindPubG2User(),
		bc.getBlindCertificate(), bc.getBlindGenerator(), VERIFY_BLIND_CERTIFICATE_API_URL);

	// A Blockchain call is made to check the validity of the User's certificate on
	// chain
	HFClient client = BlockhainCommunicator.prepareClient();
	Channel channel = client.getChannel(CHANNEL_NAME);
	BlockhainCommunicator.invokeBlockChain(client, CHAINCODE_NAME, CHAINCODE_FUNCTION,
		new String[] { bc.getBlindCommitment(), bc.getBlindCertificate(), bc.getBlindPubG1CP(),
			bc.getBlindPubG2User(), bc.getBlindGenerator() });
    }

    public static Key getKeys(String api) {
	String pubKey = "";
	String privKey = "";
	String json = "";
	Key userKey = new Key();
	json = HTTPClient.get(api, null, 600);
	try {
	    Map<String, Object> map = mapper.readValue(json, new TypeReference<Map<String, Object>>() {
	    });
	    privKey = map.get("priv").toString();
	    pubKey = map.get("pub").toString();
	    userKey.setPrivKey(privKey);
	    userKey.setPubKey(pubKey);
	    LOG.info(" privKey " + privKey);
	    LOG.info(" pubKey " + pubKey);
	} catch (IOException e) {
	    e.printStackTrace();
	}
	return userKey;
    }

    public static Commitment commit(String userPubKey, String age, String api)
	    throws UnsupportedEncodingException, Exception {
	Commitment com = new Commitment();
	String res = "";

	JSONObject data = new JSONObject();
	data.put("pub", userPubKey);
	data.put("age", age);
	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);
	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	String commitment = map.get("commitment").toString();
	String random = map.get("random").toString();
	com.setCommit(commitment);
	com.setRand(random);
	LOG.info("commitment " + commitment);
	LOG.info("random " + random);
	return com;
    }

    public static Signature ivSign(String commitment, String ivPrivKey, String ivPubKey, String api)
	    throws UnsupportedEncodingException, Exception {
	Signature sign = new Signature();
	String res = "";
	JSONObject data = new JSONObject();
	data.put("commitment", commitment);
	data.put("priv", ivPrivKey);
	data.put("pub", ivPubKey);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	String s = map.get("s").toString();
	String r = map.get("r").toString();
	sign.setSign(s);
	sign.setRand(r);
	LOG.info(" IV signature " + s);
	LOG.info(" IV random " + r);
	return sign;
    }

    public static String ivVerify(String s, String r, String ivPubKey, String commitment, String api)
	    throws UnsupportedEncodingException, Exception {
	String isValid = "";
	String res = "";
	JSONObject data = new JSONObject();
	data.put("s", s);
	data.put("r", r);
	data.put("pub", ivPubKey);
	data.put("commitment", commitment);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	isValid = map.get("verify").toString();
	LOG.info(" isValid " + isValid);

	return isValid;
    }

    public static ZKP generateZKP(String secret, String userPubKey, String api)
	    throws UnsupportedEncodingException, Exception {
	ZKP zk = new ZKP();

	String res = "";
	JSONObject data = new JSONObject();
	data.put("secret", secret);
	data.put("pub", userPubKey);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	// log.info(res.toString());
	String A = map.get("A").toString();
	String t = map.get("t").toString();
	String pubSecret = map.get("pubSecret").toString();

	zk.setA(A);
	zk.setT(t);
	zk.setPubSecret(pubSecret);

	LOG.info(" ZK A " + A);
	LOG.info(" ZK T " + t);
	LOG.info(" ZK pubSecret " + pubSecret);
	return zk;
    }

    public static String verifyProof(String A, String t, String userPubKey, String pubSecret, String api)
	    throws UnsupportedEncodingException, Exception {
	String isValid = "";
	String res = "";
	JSONObject data = new JSONObject();
	data.put("A", A);
	data.put("t", t);
	data.put("pub", userPubKey);
	data.put("pubSecret", pubSecret);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	isValid = map.get("verify").toString();
	LOG.info(" isValid " + isValid);

	return res;
    }

    public static KeyPairing generateKeyPairing(String api) {
	String json = "";
	KeyPairing kp = new KeyPairing();
	json = HTTPClient.get(api, null, 600);
	try {
	    Map<String, Object> map = mapper.readValue(json, new TypeReference<Map<String, Object>>() {
	    });
	    kp.setPrivPairing(map.get("priv").toString());
	    kp.setG1Pub(map.get("g1Pub").toString());
	    kp.setG2Pub(map.get("g2Pub").toString());

	    LOG.info(" priv " + kp.getPrivPairing());
	    LOG.info(" g1Pub " + kp.getG1Pub());
	    LOG.info(" g2Pub " + kp.getG2Pub());
	} catch (IOException e) {
	    // TODO Auto-generated catch block
	    e.printStackTrace();
	}
	return kp;
    }

    public static String generateCertificate(String commitment, String privCP, String pubG2User, String api)
	    throws UnsupportedEncodingException, Exception {
	String cert = "";
	String res = "";
	JSONObject data = new JSONObject();
	data.put("commitment", commitment);
	data.put("privCP", privCP);
	data.put("pubG2User", pubG2User);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	cert = map.get("certificate").toString();
	LOG.info(" cert " + cert);

	return cert;
    }

    public static String verifyCertificate(String commitment, String certificate, String pubG1CP, String pubG2User,
	    String api) throws UnsupportedEncodingException, Exception {
	String isValid = "";
	String res = "";
	JSONObject data = new JSONObject();
	data.put("commitment", commitment);
	data.put("certificate", certificate);
	data.put("pubG1CP", pubG1CP);
	data.put("pubG2User", pubG2User);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	isValid = map.get("verify").toString();
	LOG.info(" isValid " + isValid);

	return isValid;
    }

    public static BlindCertificate blindCertificate(String commitment, String certificate, String pubG1CP,
	    String pubG2User, String privUser, String api) throws Exception {
	BlindCertificate bc = new BlindCertificate();

	String res = "";
	JSONObject data = new JSONObject();
	data.put("commitment", commitment);
	data.put("certificate", certificate);
	data.put("pubG1CP", pubG1CP);
	data.put("pubG2User", pubG2User);
	data.put("privUser", privUser);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	bc.setBlindCommitment(map.get("blindCommitment").toString());
	bc.setBlindCertificate(map.get("blindCertificate").toString());
	bc.setBlindPubG1CP(map.get("blindPubG1CP").toString());
	bc.setBlindPubG2User(map.get("blindPubG2User").toString());
	bc.setBlindPrivUser(map.get("blindPrivUser").toString());
	bc.setBlindGenerator(map.get("blindGenerator").toString());
	bc.setBlindFactor(map.get("blindFactor").toString());

	LOG.info(" bc.getBlindCommitment() : " + bc.getBlindCommitment());
	LOG.info(" bc.getBlindCertificate() : " + bc.getBlindCertificate());
	LOG.info(" bc.getBlindPubG1CP() : " + bc.getBlindPubG1CP());
	LOG.info(" bc.getBlindPubG2User() : " + bc.getBlindPubG2User());
	LOG.info(" bc.getBlindPrivUser() : " + bc.getBlindPrivUser());
	LOG.info(" bc.getBlindGenerator() : " + bc.getBlindGenerator());
	LOG.info(" bc.getBlindFactor() : " + bc.getBlindFactor());

	return bc;
    }

    public static String verifyBlindCertificate(String blindCommitment, String blindPubG1CP, String blindPubG2User,
	    String blindCertificate, String blindGenerator, String api) throws UnsupportedEncodingException, Exception {
	String isValid = "";
	String res = "";
	JSONObject data = new JSONObject();
	data.put("blindCommitment", blindCommitment);
	data.put("blindPubG1CP", blindPubG1CP);
	data.put("blindPubG2User", blindPubG2User);
	data.put("blindCertificate", blindCertificate);
	data.put("blindGenerator", blindGenerator);

	res = HTTPClient.post(api, new StringEntity(data.toJSONString(), "application/json", "UTF-8"), 600);

	Map<String, Object> map = mapper.readValue(res, new TypeReference<Map<String, Object>>() {
	});
	isValid = map.get("verify").toString();
	LOG.info(" isValid " + isValid);

	return isValid;
    }

}
