package irt.systemx.blockchain;

import java.io.IOException;
import org.apache.http.ParseException;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpDelete;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.apache.log4j.Logger;

public class HTTPClient {

    private static final Logger LOGGER = Logger.getLogger(HTTPClient.class.getName());

    public enum Method {
	GET, POST
    };

    private String host;
    private int port;

    public HTTPClient(String host, int port) {
	this.host = host;
	this.port = port;
    }

    public String getRootUrl() {
	return "http://" + host + ":" + port;
    }

    public static String get(String url) {
	return get(url, null, 30);
    }

    public static String get(String url, String jwt, int delay) {
	String response = "";
	try {
	    RequestConfig config = RequestConfig.custom().setSocketTimeout(delay * 1000).setConnectTimeout(delay * 1000)
		    .setConnectionRequestTimeout(delay * 1000).setStaleConnectionCheckEnabled(true).build();
	    CloseableHttpClient httpClient = HttpClientBuilder.create().setDefaultRequestConfig(config).build();
	    HttpGet getMethod = new HttpGet(url);
	    if (jwt != null)
		getMethod.addHeader("Authorization", "Bearer " + jwt);
	    try {
		CloseableHttpResponse httpResponse = httpClient.execute(getMethod);
		response = EntityUtils.toString(httpResponse.getEntity(), "UTF-8");
	    } finally {
		getMethod.releaseConnection();
	    }
	} catch (IOException | ParseException e) {
	    LOGGER.error("Query error", e);
	}
	return response;
    }

    public static String post(String url, StringEntity query, int delay) throws Exception {
	String response = "";
	Boolean isOk = true;

	try {
	    CloseableHttpClient httpClient = HttpClientBuilder.create().build();
	    HttpPost postMethod = new HttpPost(url);
	    postMethod.setHeader("Connection", "close");
	    try {
		postMethod.setEntity(query);
		CloseableHttpResponse httpResponse = httpClient.execute(postMethod);
		if (httpResponse.getStatusLine().getStatusCode() != 200) {
		    isOk = false;
		}
		response = EntityUtils.toString(httpResponse.getEntity(), "UTF-8");
	    } finally {
		postMethod.releaseConnection();
	    }
	} catch (IOException | ParseException e) {
	    LOGGER.error("Query error", e);
	}

	if (!isOk) {
	    throw new Exception(response == null ? "NULL" : response);
	}

	return response;
    }

    public static String delete(String url) {
	String response = "";
	try {
	    CloseableHttpClient httpClient = HttpClients.createDefault();
	    HttpDelete deleteMethod = new HttpDelete(url);
	    try {
		CloseableHttpResponse httpResponse = httpClient.execute(deleteMethod);
		response = EntityUtils.toString(httpResponse.getEntity());
	    } finally {
		deleteMethod.releaseConnection();
	    }
	} catch (IOException | ParseException e) {
	    LOGGER.error("Query error", e);
	}
	return response;
    }
}
