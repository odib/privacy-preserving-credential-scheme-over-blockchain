package irt.systemx.blockchain;

import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.EncodedKeySpec;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.PKCS8EncodedKeySpec;
import java.util.Collection;
import java.util.LinkedList;
import java.util.Properties;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CompletionException;
import java.util.concurrent.TimeUnit;

import org.apache.commons.codec.binary.Base64;
import org.apache.commons.io.FileUtils;
import org.apache.log4j.Logger;
import org.hyperledger.fabric.sdk.BlockEvent;
import org.hyperledger.fabric.sdk.ChaincodeID;
import org.hyperledger.fabric.sdk.ChaincodeResponse.Status;
import org.hyperledger.fabric.sdk.Channel;
import org.hyperledger.fabric.sdk.Enrollment;
import org.hyperledger.fabric.sdk.HFClient;
import org.hyperledger.fabric.sdk.Orderer;
import org.hyperledger.fabric.sdk.Peer;
import org.hyperledger.fabric.sdk.ProposalResponse;
import org.hyperledger.fabric.sdk.QueryByChaincodeRequest;
import org.hyperledger.fabric.sdk.TransactionProposalRequest;
import org.hyperledger.fabric.sdk.exception.CryptoException;
import org.hyperledger.fabric.sdk.exception.InvalidArgumentException;
import org.hyperledger.fabric.sdk.exception.ProposalException;
import org.hyperledger.fabric.sdk.exception.TransactionException;
import org.hyperledger.fabric.sdk.security.CryptoSuite;
import org.hyperledger.fabric_ca.sdk.HFCAClient;
import org.hyperledger.fabric_ca.sdk.RegistrationRequest;

import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser.Feature;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;

public class BlockhainCommunicator {

    private static final Logger LOG = Logger.getLogger(BlockhainCommunicator.class);
    private static String certificateName = "";
    private static final String CA_URL = "http://127.0.0.1:7054";
    private static final String PEER1_URL = "grpc://127.0.0.1:7151"; 
    private static final String ORDERER_URL = "grpc://127.0.0.1:7050";
    private static final String CHANNEL_NAME = "mychannel";
    private static final String PEER1_NAME = "peer1-Org1"; 
    private static final String ORDERER_NAME = "orderer.example.com";  
    private static final String ORG1_NAME = "org1"; 
    private static final String ADMIN_NAME = "admin"; 
    private static final String ORG1_MSP = "Org1MSP"; 
    private static final String ADMIN_LOGIN = "rca-org0-admin"; 
    private static final String ADMIN_PASSWORD= "rca-org0-adminpw"; 
    
	

    
    public static HFClient prepareClient() throws Exception {
	certificateName = "user-certificate" + System.currentTimeMillis();
	HFCAClient caClient = getHfCaClient(CA_URL, null);
	AppUser admin = getAdmin(caClient, caClient.getCryptoSuite());
	AppUser appUser = getUser(caClient, admin, certificateName, caClient.getCryptoSuite());
	HFClient client = getHfClient();
	client.setUserContext(appUser);
	getChannel(client);
	return client;
    }

    public static Collection<ProposalResponse> queryBlockChain(HFClient client, String chainCodeName,
	    String fonctionName, String[] argsList) throws ProposalException, InvalidArgumentException {
	ChaincodeID ccId = ChaincodeID.newBuilder().setName(chainCodeName).build();
	Channel channel = client.getChannel(CHANNEL_NAME);
	QueryByChaincodeRequest qpr2 = client.newQueryProposalRequest();
	qpr2.setChaincodeID(ccId);
	qpr2.setFcn(fonctionName);
	qpr2.setArgs(argsList);
	Collection<ProposalResponse> res2 = channel.queryByChaincode(qpr2);
	// display response
	for (ProposalResponse pres : res2) {
	    String stringResponse = new String(pres.getChaincodeActionResponsePayload());
	    LOG.info(" stringResponse " + stringResponse + " Pres.getTransactionID()" + pres.getTransactionID());
	}
	return res2;
    }

    public static void invokeBlockChain(HFClient client, String chainCodeName, String fonctionName, String[] argsList) {

	ChaincodeID chaincodeID = ChaincodeID.newBuilder().setName(chainCodeName).build();
	Channel channel = client.getChannel(CHANNEL_NAME);
	try {
	    Collection<ProposalResponse> successful = new LinkedList<>();
	    Collection<ProposalResponse> failed = new LinkedList<>();
	    TransactionProposalRequest tpr = client.newTransactionProposalRequest();
	    tpr.setChaincodeID(chaincodeID);
	    tpr.setFcn(fonctionName);
	    tpr.setArgs(argsList);
	    tpr.setProposalWaitTime(100000);
	    Collection<ProposalResponse> invokePropResp = channel.sendTransactionProposal(tpr);
	    for (ProposalResponse response : invokePropResp) {
		if (response.getStatus() == Status.SUCCESS) {
		    successful.add(response);
		} else {
		    failed.add(response);
		}
	    }

	    if (failed.size() > 0) {
		ProposalResponse firstTransactionProposalResponse = failed.iterator().next();

		throw new ProposalException(
			"Not enough endorsers for invoke(move a,b,%s):%d endorser error:%s. Was verified:%b"
				+ firstTransactionProposalResponse.getStatus().getStatus()
				+ firstTransactionProposalResponse.getMessage()
				+ firstTransactionProposalResponse.isVerified());

	    }
	    CompletableFuture<BlockEvent.TransactionEvent> txs = channel.sendTransaction(successful);
	    BlockEvent.TransactionEvent event = txs.get(60, TimeUnit.SECONDS);
	    if (event.isValid()) {
		LOG.info("Transacion tx: " + event.getTransactionID() + " is completed.");
	    } else {
		LOG.error("Transaction tx: " + event.getTransactionID() + " is invalid.");
	    }

	} catch (Exception e) {
	    LOG.debug("Error while sending a transaction ... ");
	    throw new CompletionException(e);
	}
    }

    static Channel getChannel(HFClient client) throws InvalidArgumentException, TransactionException {
	Peer peer = client.newPeer(PEER1_NAME, PEER1_URL);
	Orderer orderer = client.newOrderer(ORDERER_NAME, ORDERER_URL);
	Channel channel = client.newChannel(CHANNEL_NAME);
	channel.addPeer(peer);
	channel.addOrderer(orderer);
	channel.initialize();
	return channel;
    }

    static HFClient getHfClient() throws Exception {
	CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();
	HFClient client = HFClient.createNewInstance();
	client.setCryptoSuite(cryptoSuite);
	return client;
    }

    static AppUser getUser(HFCAClient caClient, AppUser registrar, String userId, CryptoSuite cryptosuite)
	    throws Exception {
	AppUser appUser = tryDeserialize(userId, cryptosuite);
	if (appUser == null) {
	    RegistrationRequest rr = new RegistrationRequest(userId, ORG1_NAME);
	    String enrollmentSecret = caClient.register(rr, registrar);
	    Enrollment enrollment = caClient.enroll(userId, enrollmentSecret);
	    appUser = new AppUser(userId, ORG1_NAME, ORG1_MSP, enrollment);
	    serialize(appUser, cryptosuite);
	}
	return appUser;
    }

    static AppUser getAdmin(HFCAClient caClient, CryptoSuite cryptosuite) throws Exception {
	AppUser admin = tryDeserialize(ADMIN_NAME, cryptosuite);
	if (admin == null) {
	    Enrollment adminEnrollment = caClient.enroll(ADMIN_LOGIN, ADMIN_PASSWORD);
	    admin = new AppUser(ADMIN_LOGIN, ORG1_NAME, ORG1_MSP, adminEnrollment);
	    serialize(admin, cryptosuite);
	}
	return admin;
    }

    static HFCAClient getHfCaClient(String caUrl, Properties caClientProperties) throws Exception {
	CryptoSuite cryptoSuite = CryptoSuite.Factory.getCryptoSuite();
	HFCAClient caClient = HFCAClient.createNewInstance(caUrl, caClientProperties);
	caClient.setCryptoSuite(cryptoSuite);
	return caClient;
    }

    static void serialize(AppUser appUser, CryptoSuite cryptosuite) throws IOException {

	PrivateKey key = appUser.getEnrollment().getKey();
	String Algo = key.getAlgorithm();

	String privateKeystring = new String(Base64.encodeBase64(key.getEncoded()));

	String userJson = "{\"name\":\"" + appUser.getName() + "\",\"mspid\":\"" + appUser.getMspId() + "\",\"roles\":"
		+ appUser.getRoles() + ",\"affiliation\":\"" + appUser.getAffiliation() + "\",\"algo\":\"" + Algo
		+ "\",\"enrollmentSecret\":\"\",\"enrollment\":{\"signingIdentity\":\"\n" + privateKeystring
		+ "\", \"identity\":{ \"certificate\":\"" + appUser.getEnrollment().getCert() + "\"}}}";

	try (PrintWriter out = new PrintWriter(appUser.getName() + ".jso")) {
	    out.println(userJson);
	}

    }

    static AppUser tryDeserialize(String name, CryptoSuite cryptosuite) throws Exception {
	if (Files.exists(Paths.get(name + ".jso"))) {
	    return deserialize(name, cryptosuite);
	}
	return null;
    }

    static AppUser deserialize(String name, CryptoSuite cryptosuite) {
	String str = null;
	try {
	    str = FileUtils.readFileToString(new File(Paths.get(name + ".jso").toString()), "UTF-8");
	} catch (IOException e1) {
	    e1.printStackTrace();
	}

	ObjectMapper mapper = new ObjectMapper();
	User myuser;
	AppUser appUser = null;
	mapper.configure(Feature.ALLOW_UNQUOTED_CONTROL_CHARS, true);

	PublicKey publicKey;
	PrivateKey privatekey;
	KeyFactory keyFactory;
	try {
	    myuser = mapper.readValue(str, User.class);
	    keyFactory = KeyFactory.getInstance(myuser.getAlgo());
	    EncodedKeySpec privateKeySpec = new PKCS8EncodedKeySpec(
		    Base64.decodeBase64(myuser.getEnrollment().getSigningIdentity()));
	    privatekey = keyFactory.generatePrivate(privateKeySpec);

	    publicKey = cryptosuite
		    .bytesToCertificate(
			    myuser.getEnrollment().getIdentity().getCertificate().getBytes(StandardCharsets.UTF_8))
		    .getPublicKey();
	    KeyPair keypair = new KeyPair(publicKey, privatekey);

	    appUser = new AppUser(myuser.getName(), myuser.getAffiliation(), myuser.getMspid(),
		    new EnrollementPoc(keypair, myuser.getEnrollment().getIdentity().getCertificate()));
	} catch (NoSuchAlgorithmException e) {
	    e.printStackTrace();
	} catch (InvalidKeySpecException e) {
	    e.printStackTrace();
	} catch (CryptoException e) {
	    e.printStackTrace();
	} catch (JsonParseException e) {
	    e.printStackTrace();
	} catch (JsonMappingException e) {
	    e.printStackTrace();
	} catch (IOException e) {
	    e.printStackTrace();
	}

	return appUser;
    }

}
