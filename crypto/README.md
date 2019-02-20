# Java Orchestrator

This Java code aims at implementing the workflow of a novel user-centric and privacy-preserving credential.
Along with this module, a blockchain and GO modules exist.  
The blockchain module mainly consists of anonymously verifying the user attributes via a chaincode, while the GO one is primarily used to implement all the crypto tasks.
This code can therefore be seen as an orchestrator that executes the right Crypto services written in GO, and the right chaincode function according to our protocol in order to make on chain anonymous attribute verification.


## Requirements 

To run and manipulate this code, the following elements have to be installed
- A PC with OS Linux, MAC or Windows 
- Java version >= 1.7 
- JAVA IDE like eclipse with MAVEN installed


## How to run this code 

To run this code, do the following
- #1 Using a JAVA IDE like Eclipse, Open the Class `MainProtocol.java` located in the package `irt.systemx.main`
- #2 Execute the code inside the class as a Java Application 

As previously said, this Java module cannot properly function unless the blockchain and GO modules are up. 


## Contributors

- Omar DIB | IRT SystemX (https://www.irt-systemx.fr/en/)
- Clément HUYART | ERCOM (https://www.ercom.fr/)

